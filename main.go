package main

import "gotohellava/ava"

const (
	courseURL = "https://ava.ufms.br/course/view.php?id=60741"
	username  = "igor.lomba"
	password  = "password"
)

func main() {
	// TODO transform this into a CLI
	ava.RodVisit(courseURL+"#section-1", username, password)
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
