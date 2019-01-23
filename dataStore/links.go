package dataStore

import "fmt"

/*
*
* CREATE Link
 */
func (store *DBStore) AddLink(link string, statusCode int, pageTitle string) error {

	// This should update last_updated if a conflict arises
	// This should also add/update status code
	_, err := store.DB.Exec(`INSERT INTO sites (link, status_code, title) VALUES ($1, $2, $3) 
		ON CONFLICT (link) DO UPDATE
		SET (last_updated, status_code) = (now(), $4);`,
		link, statusCode, pageTitle, statusCode)
	if err != nil {
		panic(err)
		return err
	}
	fmt.Println("ADDED URL: ", link)
	fmt.Println("----------------------------------------")
	return err
}

func (store *DBStore) RetrieveLastUsedLink() (string, error) {
	// order by last_updated and return link with youngest date
	row, err := store.DB.Query(`SELECT link FROM sites ORDER BY last_updated DESC LIMIT 1`)
	if err != nil {
		fmt.Println(err)
		return "https://en.wikipedia.org/wiki/Main_Page", err
	}
	defer row.Close()

	returnLink := ""
	for row.Next() {
		err = row.Scan(&returnLink)
		if err != nil {
			fmt.Println(err)
			return "https://en.wikipedia.org/wiki/Main_Page", err
		}
	}

	return returnLink, nil
}
