package types

import (
	"fmt"
)

const (
	ViewMenu         = "templates/menu.html"
	ViewHeader       = "templates/header.html"
	ViewSidebar      = "templates/sidebar.html"
	ViewContact      = "templates/details.html"
	ViewErr          = "templates/error.html"
	ViewNavbar       = "templates/navbar.html"
	ViewMenuConstant = "templates/navbar_constant.html"
	ViewMenuUs       = "templates/navbarus.html"
	ViewNavButtons   = "templates/navbuttons.html"
	ViewCustSearch   = "templates/custsearch.html"
	ViewLoginInsert  = "templates/logininsert.html"
	KListLimit       = 20

	KLoginFormID = 789
)

const (
	KFieldNavCount = iota
	KFieldNavPage
	KFieldNavLink
)

type (
	CheckBoxRec struct {
		Title    string
		Value    string
		Selected bool
	}

	HeaderRecord struct {
		Title    string
		Name     string
		Loggedin bool
		God      bool
		NewStyle bool
		Scripts  []string
		Styles   []string
	}

	Session struct {
		LoggedIn   bool
		IsGod      bool
		UserNumber int
	}

	KeyRecord struct {
		Symbology string
		Key       string
	}

	MainPageRecord struct {
		UserName   string
		Password   string
		Remember   bool
		Checkfield string
		Speakers   int
		Talks      int
		Locations  int
		Groups     int
		Places     int
	}

	NavButtonRecord struct {
		HasNav   bool
		HasNext  bool
		HasPrev  bool
		PrevLink string
		NextLink string
	}

	NavItem struct {
		Text  string
		Title string
		Link  string
	}

	TableRow struct {
		First  string
		Second string
		Third  string
		Fourth string
		Fifth  string
		Sixth  string
	}

	M  map[int]interface{}
	MS map[string]interface{}
	S  []MS
)

func (nr *NavButtonRecord) SetNavButtons(data M) {
	nr.HasNav = false
	nr.HasNext = false
	nr.HasPrev = false

	count, ok := data[KFieldNavCount].(int)
	if !ok {
		return
	}

	pageNumber, ok := data[KFieldNavPage].(int)
	if !ok {
		return
	}

	link, ok := data[KFieldNavLink].(string)
	if !ok {
		return
	}

	if count >= KListLimit {
		nr.HasNav = true
		nr.HasNext = true
		nr.NextLink = fmt.Sprintf("%s/%d", link, pageNumber+1)
	}
	if pageNumber > 0 {
		nr.HasNav = true
		nr.HasPrev = true
		if pageNumber == 1 {
			nr.PrevLink = link
		} else {
			nr.PrevLink = fmt.Sprintf("%s/%d", link, pageNumber-1)
		}
	}
}
