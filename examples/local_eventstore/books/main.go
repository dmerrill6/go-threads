package main

import (
	"io/ioutil"
	"os"

	core "github.com/textileio/go-threads/core/store"
	"github.com/textileio/go-threads/store"
)

type book struct {
	ID     core.EntityID
	Title  string
	Author string
	Meta   bookStats
}

type bookStats struct {
	TotalReads int
	Rating     float64
}

func main() {
	s, clean := createMemStore()
	defer clean()

	model, err := s.Register("Book", &book{})
	checkErr(err)

	// Bootstrap the model with some books: two from Author1 and one from Author2
	{
		// Create a two books for Author1
		book1 := &book{ // Notice ID will be autogenerated
			Title:  "Title1",
			Author: "Author1",
			Meta:   bookStats{TotalReads: 100, Rating: 3.2},
		}
		book2 := &book{
			Title:  "Title2",
			Author: "Author1",
			Meta:   bookStats{TotalReads: 150, Rating: 4.1},
		}
		checkErr(model.Create(book1, book2)) // Note you can create multiple books at the same time (variadic)

		// Create book for Author2
		book3 := &book{
			Title:  "Title3",
			Author: "Author2",
			Meta:   bookStats{TotalReads: 500, Rating: 4.9},
		}
		checkErr(model.Create(book3))
	}

	// Query all the books
	{
		var books []*book
		err := model.Find(&books, &store.Query{})
		checkErr(err)
		if len(books) != 3 {
			panic("there should be three books")
		}
	}

	// Query the books from Author2
	{
		var books []*book
		err := model.Find(&books, store.Where("Author").Eq("Author1"))
		checkErr(err)
		if len(books) != 2 {
			panic("Author1 should have two books")
		}
	}

	// Query with nested condition
	{
		var books []*book
		err := model.Find(&books, store.Where("Meta.TotalReads").Eq(100))
		checkErr(err)
		if len(books) != 1 {
			panic("There should be one book with 100 total reads")
		}
	}

	// Query book by two conditions
	{
		var books []*book
		err := model.Find(&books, store.Where("Author").Eq("Author1").And("Title").Eq("Title2"))
		checkErr(err)
		if len(books) != 1 {
			panic("Author1 should have only one book with Title2")
		}
	}

	// Query book by OR condition
	{
		var books []*book
		err := model.Find(&books, store.Where("Author").Eq("Author1").Or(store.Where("Author").Eq("Author2")))
		checkErr(err)
		if len(books) != 3 {
			panic("Author1 & Author2 have should have 3 books in total")
		}
	}

	// Sorted query
	{
		var books []*book
		// Ascending
		err := model.Find(&books, store.Where("Author").Eq("Author1").OrderBy("Meta.TotalReads"))
		checkErr(err)
		if books[0].Meta.TotalReads != 100 || books[1].Meta.TotalReads != 150 {
			panic("books aren't ordered asc correctly")
		}
		// Descending
		err = model.Find(&books, store.Where("Author").Eq("Author1").OrderByDesc("Meta.TotalReads"))
		checkErr(err)
		if books[0].Meta.TotalReads != 150 || books[1].Meta.TotalReads != 100 {
			panic("books aren't ordered desc correctly")
		}
	}

	// Query, Update, and Save
	{
		var books []*book
		err := model.Find(&books, store.Where("Title").Eq("Title3"))
		checkErr(err)

		// Modify title
		book := books[0]
		book.Title = "ModifiedTitle"
		_ = model.Save(book)
		err = model.Find(&books, store.Where("Title").Eq("Title3"))
		checkErr(err)
		if len(books) != 0 {
			panic("Book with Title3 shouldn't exist")
		}

		// Delete it
		err = model.Find(&books, store.Where("Title").Eq("ModifiedTitle"))
		checkErr(err)
		if len(books) != 1 {
			panic("Book with ModifiedTitle should exist")
		}
		_ = model.Delete(books[0].ID)
		err = model.Find(&books, store.Where("Title").Eq("ModifiedTitle"))
		checkErr(err)
		if len(books) != 0 {
			panic("Book with ModifiedTitle shouldn't exist")
		}
	}
}

func createMemStore() (*store.Store, func()) {
	dir, err := ioutil.TempDir("", "")
	checkErr(err)
	ts, err := store.DefaultService(dir)
	checkErr(err)
	s, err := store.NewStore(ts, store.WithRepoPath(dir))
	checkErr(err)
	return s, func() {
		if err := ts.Close(); err != nil {
			panic(err)
		}
		_ = os.RemoveAll(dir)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
