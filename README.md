# osrs-scraper
This is a web scraper that scrapes the official [oldschool runescape highscores](https://secure.runescape.com/m=hiscore_oldschool/overall.ws?table=0&page=1) list, as there is no official API to retrieve the list of all users.

The maximum amount of players that can be trertieved is 2,000,000. This is a limit set by the developers of the website.

Each page retrieves 25 players, and thus the last page is 80,000.

## getPageContentByPageNumber (url string) []byte
This function takes in a page nunmber as a paramater, connects to that page of the oldschool runescape highscores, scrapes the page data, and retuns that data as a byte slice.

## getCleanedTableBodyData (HTMLData []byte) []byte
This function takes html data retrieved by the `getPageContentFromUrl` function, and finds the first html table body within it.

It then cleans the data by doing the following:
- replacing all `\"` with `'`. Because When reading the page data, anywhere there might be quotes, for example `class="someClassName"`, the output would be `class=/"someClassName/"`. So after cleaning, it would be: `class='someClassName'`.
- removing all `\n`.
- replacing all `\xa0` with a proper space.
- removes all `,`.

Once this is finished, it then returns the cleaned table body html.
