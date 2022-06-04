package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	uuid "github.com/satori/go.uuid"
)

type postDisplay struct {
	PostID        string
	Username      string
	PostCategory  string
	Likes         int
	Dislikes      int
	TitleText     string
	PostText      string
	CookieChecker bool
	Comments      []commentStruct
}

type commentStruct struct {
	CommentID       string
	CpostID         string
	CommentUsername string
	CommentText     string
	Likes           int
	Dislikes        int
	CookieChecker   bool
}

func newPost(userName, category, title, post string, db *sql.DB) {
	if title == "" {
		return
	}

	fmt.Println("ADDING POST")
	uuid := uuid.NewV4().String()
	_, err := db.Exec("INSERT INTO posts (postID, userName, category, likes, dislikes, title, post) VALUES (?, ?, ?, 0, 0, ?, ?)", uuid, userName, category, title, post)
	if err != nil {
		fmt.Println("Error adding new post")
		log.Fatal(err.Error())
	}
	Person.PostAdded = true

	catSlc := strings.Split(category, " ")
	feSelected := 0
	beSelected := 0
	fsSelected := 0
	//Loop through categories if any element = Animals, Travel or Movies chane accordingly
	for _, r := range catSlc {
		if r == "Animals" {
			feSelected = 1
		} else if r == "Travel" {
			beSelected = 1
		} else if r == "Movies" {
			fsSelected = 1
		}
	}

	_, errAddCats := db.Exec("INSERT INTO categories (postID, Animals, Travel, Movies) VALUES (?, ?, ?, ?)", uuid, feSelected, beSelected, fsSelected)
	if errAddCats != nil {
		fmt.Println("ERROR when adding into the category table")
	}
}

func postData(db *sql.DB) []postDisplay {
	rows, err := db.Query("SELECT postID, userName, category, likes, dislikes, title, post FROM posts")
	if err != nil {
		fmt.Println("Error selecting post data")
		log.Fatal(err.Error())
	}

	finalArray := []postDisplay{}

	for rows.Next() {

		var u postDisplay
		err := rows.Scan(
			&u.PostID,
			&u.Username,
			&u.PostCategory,
			&u.Likes,
			&u.Dislikes,
			&u.TitleText,
			&u.PostText,
		)
		u.CookieChecker = Person.CookieChecker
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}

		commentSlc := []commentStruct{}
		var tempComStruct commentStruct

		commentRow, errComs := db.Query("SELECT commentID, postID, username, commentText, likes, dislikes FROM comments WHERE postID = ?", u.PostID)
		if errComs != nil {
			fmt.Println("Error selecting comment data")
			log.Fatal(errComs.Error())
		}
		for commentRow.Next() {
			err := commentRow.Scan(
				&tempComStruct.CommentID,
				&tempComStruct.CpostID,
				&tempComStruct.CommentUsername,
				&tempComStruct.CommentText,
				&tempComStruct.Likes,
				&tempComStruct.Dislikes,
			)
			tempComStruct.CookieChecker = Person.CookieChecker
			if err != nil {
				fmt.Println("Error scanning comments")
				log.Fatal(err.Error())
			}
			fmt.Printf("\nCOMMENT STRUCT_____-------------------------------------%v\n\n", tempComStruct)
			commentSlc = append(commentSlc, tempComStruct)
		}
		u.Comments = commentSlc

		finalArray = append(finalArray, u)

		for i, j := 0, len(finalArray)-1; i < j; i, j = i+1, j-1 {
			finalArray[i], finalArray[j] = finalArray[j], finalArray[i]
		}
	}
	return finalArray
}

func LikeButton(postID string, db *sql.DB) {
	findRow, errRows := db.Query("SELECT reference FROM liketable WHERE postID = (?) AND user = (?)", postID, Person.Username)
	if errRows != nil {
		fmt.Println("SELECTING LIKE ERROR")
		log.Fatal(errRows.Error())
	}
	rounds := 0

	var check postDisplay
	for findRow.Next() {
		rounds++
		err2 := findRow.Scan(
			&check.Likes,
		)

		if err2 != nil {
			log.Fatal(err2.Error())
		}
	}

	if rounds == 0 {
		_, insertLikeErr := db.Exec("INSERT INTO liketable (user, postID, reference) VALUES (?, ?, 1)", Person.Username, postID)
		if insertLikeErr != nil {
			fmt.Println("Error when inserting into like table initially (LIKEBUTTON)")
			log.Fatal(insertLikeErr.Error())
		}

		LikeIncrease(postID, sqliteDatabase)
	} else {
		if check.Likes == 1 {
			LikeUndo(postID, sqliteDatabase)
			RefUpdate(0, postID, sqliteDatabase)
		} else if check.Likes == -1 {
			DislikeUndo(postID, sqliteDatabase)
			LikeIncrease(postID, sqliteDatabase)

			RefUpdate(1, postID, sqliteDatabase)

		} else if check.Likes == 0 {
			LikeIncrease(postID, sqliteDatabase)
			RefUpdate(1, postID, sqliteDatabase)

		}
	}
}

func RefUpdate(value int, postID string, db *sql.DB) {
	_, err2 := db.Exec("UPDATE liketable SET reference = (?) WHERE postID = (?) AND user = (?)", value, postID, Person.Username)
	if err2 != nil {
		fmt.Println("UPDATING REFERENCE ")
		log.Fatal(err2.Error())
	}
}
func CommentRefUpdate(value int, postID string, db *sql.DB) {
	_, err2 := db.Exec("UPDATE liketable SET reference = (?) WHERE commentID = (?) AND user = (?)", value, postID, Person.Username)
	if err2 != nil {
		fmt.Println("UPDATING REFERENCE ")
		log.Fatal(err2.Error())
	}
}

func LikeIncrease(postID string, db *sql.DB) {

	likes, err := db.Query("SELECT likes FROM posts WHERE postID = (?)", postID)
	if err != nil {
		fmt.Println("Error selecting likes")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for likes.Next() {
		err := likes.Scan(
			&temp.Likes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Likes++
	_, err2 := db.Exec("UPDATE posts SET likes = (?) WHERE postID = (?)", temp.Likes, postID)
	if err2 != nil {
		fmt.Println("UPDATING LIKES WHEN ROUNDS == 0")
		log.Fatal(err.Error())
	}

}

func LikeUndo(postID string, db *sql.DB) {
	likes, err := db.Query("SELECT likes FROM posts WHERE postID = (?)", postID)
	if err != nil {
		fmt.Println("Error in LIKE UNDO")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for likes.Next() {
		err := likes.Scan(
			&temp.Likes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Likes--
	_, err2 := db.Exec("UPDATE posts SET likes = (?) WHERE postID = (?)", temp.Likes, postID)
	if err2 != nil {
		fmt.Println("LIKE UNDO")
		log.Fatal(err.Error())
	}
}

func DislikeIncrease(postID string, db *sql.DB) {
	dislikes, err := db.Query("SELECT dislikes FROM posts WHERE postID = (?)", postID)
	if err != nil {
		fmt.Println("Error in DislikeIncrease")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for dislikes.Next() {
		err := dislikes.Scan(
			&temp.Dislikes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Dislikes++
	_, err2 := db.Exec("UPDATE posts SET dislikes = (?) WHERE postID = (?)", temp.Dislikes, postID)
	if err2 != nil {
		fmt.Println("UPDATING DISLIKES")
		log.Fatal(err.Error())
	}
}

func DislikeUndo(postID string, db *sql.DB) {
	dislikes, err := db.Query("SELECT dislikes FROM posts WHERE postID = (?)", postID)
	if err != nil {
		fmt.Println("Error in DislikeUndo")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for dislikes.Next() {
		err := dislikes.Scan(
			&temp.Dislikes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Dislikes--
	_, err2 := db.Exec("UPDATE posts SET dislikes = (?) WHERE postID = (?)", temp.Dislikes, postID)
	if err2 != nil {
		fmt.Println("DISLIKE UNDO")
		log.Fatal(err.Error())
	}
}

func DislikeButton(postID string, db *sql.DB) {
	fmt.Printf("\n\n--------------THE POSTID FOR THE BUTTON CLICKED IS: %v \n\n", postID)
	findRow, errRows := db.Query("SELECT reference FROM liketable WHERE postID = (?) AND user = (?)", postID, Person.Username)
	if errRows != nil {
		fmt.Println("SELECTING LIKE ERROR")
		log.Fatal(errRows.Error())
	}
	rounds := 0

	var check postDisplay
	for findRow.Next() {
		rounds++
		err := findRow.Scan(
			&check.Likes,
		)

		if err != nil {
			fmt.Println("Error in Dislike Button")
			log.Fatal(err.Error())
		}
	}

	if rounds == 0 {
		_, insertLikeErr := db.Exec("INSERT INTO liketable (user, postID, reference) VALUES (?, ?, -1)", Person.Username, postID)
		if insertLikeErr != nil {
			fmt.Println("Error when inserting into like table initially (DISLIKEBUTTON)")
			log.Fatal(insertLikeErr.Error())
		}
		DislikeIncrease(postID, sqliteDatabase)

	} else {
		if check.Likes == -1 {
			DislikeUndo(postID, sqliteDatabase)
			RefUpdate(0, postID, sqliteDatabase)
		} else if check.Likes == 1 {
			LikeUndo(postID, sqliteDatabase)
			DislikeIncrease(postID, sqliteDatabase)
			RefUpdate(-1, postID, sqliteDatabase)
		} else if check.Likes == 0 {
			DislikeIncrease(postID, sqliteDatabase)
			RefUpdate(-1, postID, sqliteDatabase)
		}
	}
}

func newComment(userName, postID, commentText string, db *sql.DB) {
	if commentText == "" {
		return
	}

	fmt.Println("ADDING Comment")
	uuid := uuid.NewV4().String()
	_, err := db.Exec("INSERT INTO comments (commentID, postID, userName, commentText, likes, dislikes) VALUES (?, ?, ?, ?, 0, 0)", uuid, postID, userName, commentText)
	if err != nil {
		fmt.Println("ERROR ADDING COMMENT TO THE TABLE")
		log.Fatal(err.Error())
	}
	Person.PostAdded = true

}

func CommentLikeButton(postID string, db *sql.DB) {
	findRow, errRows := db.Query("SELECT reference FROM liketable WHERE commentID = (?) AND user = (?)", postID, Person.Username)
	if errRows != nil {
		fmt.Println("SELECTING LIKE ERROR")
		log.Fatal(errRows.Error())
	}
	rounds := 0

	var check postDisplay
	for findRow.Next() {
		rounds++
		err2 := findRow.Scan(
			&check.Likes,
		)

		if err2 != nil {
			log.Fatal(err2.Error())
		}
	}

	if rounds == 0 {
		_, insertLikeErr := db.Exec("INSERT INTO liketable (user, commentID, reference) VALUES (?, ?, 1)", Person.Username, postID)
		if insertLikeErr != nil {
			fmt.Println("Error when inserting into like table initially (LIKEBUTTON)")
			log.Fatal(insertLikeErr.Error())
		}

		CommentLikeIncrease(postID, sqliteDatabase)
	} else {

		fmt.Printf("\n ------------------------------------------------------------------REFERENCE is equal to: %v", check.Likes)
		if check.Likes == 1 {
			CommentLikeUndo(postID, sqliteDatabase)
			CommentRefUpdate(0, postID, sqliteDatabase)
		} else if check.Likes == -1 {
			CommentDislikeUndo(postID, sqliteDatabase)
			CommentLikeIncrease(postID, sqliteDatabase)

			CommentRefUpdate(1, postID, sqliteDatabase)

		} else if check.Likes == 0 {
			CommentLikeIncrease(postID, sqliteDatabase)
			CommentRefUpdate(1, postID, sqliteDatabase)

		}
	}
}

func CommentLikeIncrease(postID string, db *sql.DB) {

	likes, err := db.Query("SELECT likes FROM comments WHERE commentID = (?)", postID)
	if err != nil {
		fmt.Println("Error selecting likes")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for likes.Next() {
		err := likes.Scan(
			&temp.Likes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}
	fmt.Printf("CURRENT COMMENT LIKES: %v \n", temp.Likes)

	temp.Likes++
	fmt.Printf("\n INCREAED COMMENT LIKES: %v !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!111\n", temp.Likes)
	_, err2 := db.Exec("UPDATE comments SET likes = (?) WHERE commentID = (?)", temp.Likes, postID)
	if err2 != nil {
		fmt.Println("UPDATING LIKES WHEN ROUNDS == 0")
		log.Fatal(err.Error())
	}

}

func CommentLikeUndo(postID string, db *sql.DB) {
	likes, err := db.Query("SELECT likes FROM comments WHERE commentID = (?)", postID)
	if err != nil {
		fmt.Println("Error in LIKE UNDO")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for likes.Next() {
		err := likes.Scan(
			&temp.Likes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Likes--
	_, err2 := db.Exec("UPDATE comments SET likes = (?) WHERE commentID = (?)", temp.Likes, postID)
	if err2 != nil {
		fmt.Println("LIKE UNDO")
		log.Fatal(err.Error())
	}
}

func CommentDislikeUndo(postID string, db *sql.DB) {
	dislikes, err := db.Query("SELECT dislikes FROM comments WHERE commentID = (?)", postID)
	if err != nil {
		fmt.Println("Error in DislikeUndo")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for dislikes.Next() {
		err := dislikes.Scan(
			&temp.Dislikes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Dislikes--
	_, err2 := db.Exec("UPDATE comments SET dislikes = (?) WHERE commentID = (?)", temp.Dislikes, postID)
	if err2 != nil {
		fmt.Println("DISLIKE UNDO")
		log.Fatal(err.Error())
	}
}

func CommentDislikeIncrease(postID string, db *sql.DB) {
	dislikes, err := db.Query("SELECT dislikes FROM comments WHERE commentID = (?)", postID)
	if err != nil {
		fmt.Println("Error in DislikeIncrease")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for dislikes.Next() {
		err := dislikes.Scan(
			&temp.Dislikes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Dislikes++
	_, err2 := db.Exec("UPDATE comments SET dislikes = (?) WHERE commentID = (?)", temp.Dislikes, postID)
	if err2 != nil {
		fmt.Println("UPDATING DISLIKES")
		log.Fatal(err.Error())
	}
}

func CommentDislikeButton(postID string, db *sql.DB) {
	findRow, errRows := db.Query("SELECT reference FROM liketable WHERE commentID = (?) AND user = (?)", postID, Person.Username)
	if errRows != nil {
		fmt.Println("SELECTING LIKE ERROR")
		log.Fatal(errRows.Error())
	}
	rounds := 0

	var check postDisplay
	for findRow.Next() {
		rounds++
		err := findRow.Scan(
			&check.Likes,
		)

		if err != nil {
			fmt.Println("Error in Dislike Button")
			log.Fatal(err.Error())
		}
	}

	if rounds == 0 {
		_, insertLikeErr := db.Exec("INSERT INTO liketable (user, commentID, reference) VALUES (?, ?, -1)", Person.Username, postID)
		if insertLikeErr != nil {
			fmt.Println("Error when inserting into like table initially (DISLIKEBUTTON)")
			log.Fatal(insertLikeErr.Error())
		}
		CommentDislikeIncrease(postID, sqliteDatabase)

	} else {
		if check.Likes == -1 {
			CommentDislikeUndo(postID, sqliteDatabase)
			CommentRefUpdate(0, postID, sqliteDatabase)
		} else if check.Likes == 1 {
			CommentLikeUndo(postID, sqliteDatabase)
			CommentDislikeIncrease(postID, sqliteDatabase)
			CommentRefUpdate(-1, postID, sqliteDatabase)
		} else if check.Likes == 0 {
			CommentDislikeIncrease(postID, sqliteDatabase)
			CommentRefUpdate(-1, postID, sqliteDatabase)
		}
	}
}

func PostGetter(postIDSlc []string, db *sql.DB) []postDisplay {
	finalArray := []postDisplay{}
	for _, r := range postIDSlc {
		rows, errDetails := db.Query("SELECT postID, userName, category, likes, dislikes, title, post FROM posts WHERE postID = (?)", r)
		if errDetails != nil {
			fmt.Println("ERROR when selecting the information for specific posts (func POSTGETTER)")
			log.Fatal(errDetails.Error())
		}

		for rows.Next() {
			var postDetails postDisplay
			err := rows.Scan(
				&postDetails.PostID,
				&postDetails.Username,
				&postDetails.PostCategory,
				&postDetails.Likes,
				&postDetails.Dislikes,
				&postDetails.TitleText,
				&postDetails.PostText,
			)
			postDetails.CookieChecker = Person.CookieChecker
			if err != nil {
				fmt.Println("ERROR Scanning through the rows (func POSTGETTER)")
				log.Fatal(err.Error())
			}
			finalArray = append(finalArray, postDetails)
		}
	}
	return finalArray
}
