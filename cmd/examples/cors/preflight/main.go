package main

import (
	"flag"
	"log"
	"net/http"
)

// Define a string constant containing the HTML for the webpage. This consists of a <h1>
// header tag, and some Javascript which calls our POST /v1/tokens/authenticate endpoint
// and writes the response body to inside the <div id="output"></div> tag.
const html = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
</head>
<body>
	<h1>Preflight CORS</h1>
	<div id="output"></div>
	<script>
		document.addEventListener('DOMContentLoaded', function(){
			fetch("http://localhost:8000/v1/tokens/authenticate", {
				method: "POST",
				headers: {'Content-Type': 'application/json'},
				body: JSON.stringify({
					email: 'donald@example.com',
					password: 'pa55word'
				})
			}).then(
				function (response) {
					response.text().then(function (text) {
						document.getByElementId("output").innerHTML = text;
					});
				},
				function(err) {
					document.getElementById("output").innerHTML = err;
				}
			);
		});
	</script>
</body>
</html>
`

func main() {
	addr := flag.String("addr", ":9000", "server address")
	flag.Parse()

	log.Printf("Starting server on %s", *addr)
	err := http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	log.Fatal(err)
}
