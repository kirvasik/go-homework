package domain

type Movie struct {
	ID       int
	Title    string
	Year     int
	Director string
}

type FailedID struct {
	ID    int
	Error string
}

type LoadResult struct {
	Movies []Movie
	Errors []FailedID
}
