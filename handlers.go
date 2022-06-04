package main

import (
	"fmt"
	"net/http"
	"text/template"

	uuid "github.com/satori/go.uuid"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	urlError := urlError(w, r)
	if urlError {
		return
	}

	tpl := template.Must(template.ParseGlob("templates/login.html"))
	if err := tpl.Execute(w, Person); err != nil {
		http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
	}
}

func LoginResult(w http.ResponseWriter, r *http.Request) {
	urlError := urlError(w, r)
	if urlError {
		return
	}
	Person.Attempted = true
	if r.Method != "POST" && r.Method != "GET" {
		fmt.Fprint(w, r.Method+"\n")
		http.Error(w, "400 Status Bad Request", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	pass := r.FormValue("password")
	uuid := uuid.NewV4()

	if Person.Accesslevel && Person.CookieChecker {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else if Person.Accesslevel {
		Person.Attempted = false
		tpl := template.Must(template.ParseGlob("templates/login.html"))
		if err := tpl.Execute(w, Person); err != nil {
			http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
		}
	} else if ValidEmail(email, sqliteDatabase) {
		if LoginValidator(email, pass, sqliteDatabase) {

			if Person.Accesslevel {
				cookie, err := r.Cookie("1st-cookie")
				fmt.Println("cookie:", cookie, "err:", err)
				if err != nil {
					fmt.Println("cookie was not found")
					cookie = &http.Cookie{
						Name:     "1st-cookie",
						Value:    uuid.String(),
						HttpOnly: true,

						Path: "/",
					}
					http.SetCookie(w, cookie)
					CookieAdd(cookie, sqliteDatabase)
				}

			}
			Person.CookieChecker = true
			Person.Attempted = false

			x := homePageStruct{MembersPost: Person, PostingDisplay: postData(sqliteDatabase)}
			tpl := template.Must(template.ParseGlob("templates/index.html"))
			if err := tpl.Execute(w, x); err != nil {
				http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
			}
		} else {
			tpl := template.Must(template.ParseGlob("templates/login.html"))
			if err := tpl.Execute(w, Person); err != nil {
				http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
			}
		}

	} else {
		tpl := template.Must(template.ParseGlob("templates/login.html"))
		if err := tpl.Execute(w, Person); err != nil {
			http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
		}
	}
}

func registration(w http.ResponseWriter, r *http.Request) {
	urlError := urlError(w, r)
	if urlError {
		return
	}
	Person.RegistrationAttempted = false
	tpl := template.Must(template.ParseGlob("templates/register.html"))
	if err := tpl.Execute(w, Person); err != nil {
		fmt.Println(err.Error())
		http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
	}
}

func registration2(w http.ResponseWriter, r *http.Request) {
	Person.SuccessfulRegistration = false
	urlError := urlError(w, r)
	if urlError {
		return
	}
	if r.Method != "POST" && r.Method != "GET" {
		fmt.Fprint(w, r.Method+"\n")
		http.Error(w, "400 Status Bad Request", http.StatusBadRequest)
		return
	}

	userN := r.FormValue("username")
	email := r.FormValue("email")
	pass := r.FormValue("password")
	Person.RegistrationAttempted = true

	exist, _ := userExist(email, userN, sqliteDatabase)

	tpl := template.Must(template.ParseGlob("templates/register.html"))

	if exist {
		Person.FailedRegister = true
		if err := tpl.Execute(w, Person); err != nil {
			http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
		}

	} else {
		Person.SuccessfulRegistration = true
		newUser(email, userN, pass, sqliteDatabase)

		if err := tpl.Execute(w, Person); err != nil {
			http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
		}

	}

}

func Post(w http.ResponseWriter, r *http.Request) {
	urlError := urlError(w, r)
	if urlError {
		return
	}
	Person.PostAdded = false
	tpl := template.Must(template.ParseGlob("templates/newPost.html"))
	if err := tpl.Execute(w, Person); err != nil {
		http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
	}

}

func postAdded(w http.ResponseWriter, r *http.Request) {
	urlError := urlError(w, r)
	if urlError {
		return
	}
	if r.Method != "POST" && r.Method != "GET" {
		fmt.Fprint(w, r.Method+"\n")
		http.Error(w, "400 Status Bad Request", http.StatusBadRequest)
		return
	}
	FEcat := r.FormValue("Animals")
	BEcat := r.FormValue("Travel")
	FScat := r.FormValue("Movies")

	cat := FEcat + " " + BEcat + " " + FScat

	c := []rune(cat)
	category := []rune{}
	for i := 0; i < len(c); i++ {
		category = append(category, c[i])
		if c[i] == ' ' && c[i]+1 == ' ' {
			i++
		}
	}
	cat = string(category)
	title := r.FormValue("title")
	post := r.FormValue("post")
	newPost(Person.Username, cat, title, post, sqliteDatabase)

	x := homePageStruct{MembersPost: Person, PostingDisplay: postData(sqliteDatabase)}

	tpl := template.Must(template.ParseGlob("templates/index.html"))

	if err := tpl.Execute(w, x); err != nil {
		http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
	}

}

type homePageStruct struct {
	MembersPost    userDetails
	PostingDisplay []postDisplay
}

func Home(w http.ResponseWriter, r *http.Request) {
	urlError := urlError(w, r)
	if urlError {
		return
	}
	if r.Method != "POST" && r.Method != "GET" {
		fmt.Fprint(w, r.Method+"\n")
		http.Error(w, "400 Status Bad Request", http.StatusBadRequest)
		return
	}
	postNum := r.FormValue("likeBtn")
	fmt.Printf("\n LIKE BUTTON VALUE \n")
	fmt.Println(postNum)
	LikeButton(postNum, sqliteDatabase)

	dislikePostNum := r.FormValue("dislikeBtn")
	fmt.Printf("\n\n\n?????????????????????????????DISLIKE BUTTON VALUE : %v \n \n", dislikePostNum)
	DislikeButton(dislikePostNum, sqliteDatabase)

	comment := r.FormValue("commentTxt")
	commentPostID := r.FormValue("commentSubmit")
	fmt.Printf("ADDING COMMENT: %v", commentPostID)
	newComment(Person.Username, commentPostID, comment, sqliteDatabase)

	commentNum := r.FormValue("commentlikeBtn")
	fmt.Printf("\n Comment LIKE BUTTON VALUE")
	fmt.Println(commentNum)
	CommentLikeButton(commentNum, sqliteDatabase)

	commentDislike := r.FormValue("commentDislikeBtn")
	fmt.Printf("\nDISLIKE BUTTON VALUE \n")
	CommentDislikeButton(commentDislike, sqliteDatabase)

	FE := r.FormValue("FEfilter")
	BE := r.FormValue("BEfilter")
	FS := r.FormValue("FSfilter")
	MyLikes := r.FormValue("likedPosts")
	Created := r.FormValue("myPosts")

	postSlc := []postDisplay{}
	if FE == "Animals" {
		frontEndSlc := []string{}
		animals, errGetIDs := sqliteDatabase.Query("SELECT postID from categories WHERE Animals = 1")
		if errGetIDs != nil {
			fmt.Println("EEROR trying to SELECT the posts with front end ID")
		}
		for animals.Next() {
			var GetIDs commentStruct

			err := animals.Scan(
				&GetIDs.CommentID,
			)
			if err != nil {
				fmt.Println("Error Scanning through rows")
			}

			frontEndSlc = append(frontEndSlc, GetIDs.CommentID)
		}
		postSlc = PostGetter(frontEndSlc, sqliteDatabase)

	} else if BE == "Travel" {
		BackEndSlc := []string{}
		travel, errGetIDs := sqliteDatabase.Query("SELECT postID from categories WHERE Travel = 1")
		if errGetIDs != nil {
			fmt.Println("EEROR trying to SELECT the posts with front end ID")
		}
		for travel.Next() {
			var GetIDs commentStruct

			err := travel.Scan(
				&GetIDs.CommentID,
			)
			if err != nil {
				fmt.Println("Error Scanning through rows")
			}

			BackEndSlc = append(BackEndSlc, GetIDs.CommentID)
		}
		postSlc = PostGetter(BackEndSlc, sqliteDatabase)

	} else if FS == "Movies" {
		FullStackSlc := []string{}
		movies, errGetIDs := sqliteDatabase.Query("SELECT postID from categories WHERE Movies= 1")
		if errGetIDs != nil {
			fmt.Println("EEROR trying to SELECT the posts with front end ID")
		}
		for movies.Next() {
			var GetIDs commentStruct

			err := movies.Scan(
				&GetIDs.CommentID,
			)
			if err != nil {
				fmt.Println("Error Scanning through rows")
			}

			FullStackSlc = append(FullStackSlc, GetIDs.CommentID)
		}
		postSlc = PostGetter(FullStackSlc, sqliteDatabase)

	} else if MyLikes == "Liked Posts" {
		likedSlc := []string{}
		likedRows, errGetIDs := sqliteDatabase.Query("SELECT postID from liketable WHERE reference = 1 AND user = (?)", Person.Username)
		if errGetIDs != nil {
			fmt.Println("EEROR trying to SELECT the posts with front end ID")
		}
		for likedRows.Next() {
			var GetIDs commentStruct

			err := likedRows.Scan(
				&GetIDs.CommentID,
			)
			if err != nil {
				fmt.Println("Error Scanning through rows")
			}

			likedSlc = append(likedSlc, GetIDs.CommentID)
		}
		postSlc = PostGetter(likedSlc, sqliteDatabase)
	} else if Created == "My Posts" {
		myPostsSlc := []string{}
		myPostsRows, errGetIDs := sqliteDatabase.Query("SELECT postID from posts WHERE userName = (?)", Person.Username)
		if errGetIDs != nil {
			fmt.Println("EEROR trying to SELECT the posts with front end ID")
		}
		for myPostsRows.Next() {
			var GetIDs commentStruct

			err := myPostsRows.Scan(
				&GetIDs.CommentID,
			)
			if err != nil {
				fmt.Println("Error Scanning through rows")
			}

			myPostsSlc = append(myPostsSlc, GetIDs.CommentID)
		}
		postSlc = PostGetter(myPostsSlc, sqliteDatabase)

	} else {
		postSlc = postData(sqliteDatabase)
	}

	c1, err1 := r.Cookie("1st-cookie")

	if err1 == nil && !Person.Accesslevel {
		c1.MaxAge = -1
		http.SetCookie(w, c1)
	}

	c, err := r.Cookie("1st-cookie")

	if err != nil && Person.Accesslevel {
		Person.CookieChecker = false

	} else if err == nil && Person.Accesslevel {
		Person.CookieChecker = true

	} else {
		Person.CookieChecker = false
	}

	fmt.Printf("\n\n\n------------------------------------------------------------------Struct BEFORE: %v\n\n\n\n", Person)

	fmt.Printf("\n\n\n------------------------------------------------------------------Struct AFTER: %v\n\n\n\n", Person)
	x := homePageStruct{MembersPost: Person, PostingDisplay: postSlc}

	tpl := template.Must(template.ParseGlob("templates/index.html"))

	if err := tpl.Execute(w, x); err != nil {
		http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
	}
	fmt.Println("YOUR COOKIE:", c)
}

func LogOut(w http.ResponseWriter, r *http.Request) {
	urlError := urlError(w, r)
	if urlError {
		return
	}

	if !Person.Accesslevel {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	c, err := r.Cookie("1st-cookie")

	if err != nil {
		fmt.Println("Problem logging out with cookie")
	}

	c.MaxAge = -1
	http.SetCookie(w, c)

	if c.MaxAge == -1 {
		fmt.Println("Cookie deleted")
	}

	var newPerson userDetails
	Person = newPerson

	fmt.Println(Person)

	tpl := template.Must(template.ParseGlob("templates/logout.html"))

	if err := tpl.Execute(w, ""); err != nil {
		http.Error(w, "No such file or directory: Internal Server Error 500", http.StatusInternalServerError)
	}
}

func urlError(w http.ResponseWriter, r *http.Request) bool {
	p := r.URL.Path
	if !((p == "/") || (p == "/log") || (p == "/login") || (p == "/register") || (p == "/registration") || (p == "/logout") || (p == "/new-post") || (p == "/post-added")) {
		http.Error(w, "404 Status not found", http.StatusNotFound)
		return true
	}
	return false
}
