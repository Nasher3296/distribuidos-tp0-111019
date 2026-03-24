package bet

import "fmt"

type Bet struct {
	Agency    string
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    string
}

func (b Bet) ToCsvBytes() []byte {
	return []byte(fmt.Sprintf("%s,%s,%s,%s,%s,%s",
		b.Agency, b.FirstName, b.LastName, b.Document, b.Birthdate, b.Number))
}
