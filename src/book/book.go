package book

// Book represents the attributes of a book including its title, price, and whether or
// not it is in stock.
// Structure is public to all outside packages because both linkedlist and html need it.
type Book struct {
	Title   string
	Description string
	Price   float64
	InStock bool
	Stars   int
}

// NewBook returns a pointer to a Book object with givem title, price, and inStock values.
func NewBook(title string, description string, price float64, inStock bool, stars int) *Book {
	return &Book{title, description, price, inStock, stars}
}