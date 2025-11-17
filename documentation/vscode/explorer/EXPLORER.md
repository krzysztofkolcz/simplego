Przejście do explorera:
Ctrl + 0

Przejście do edytora
Ctrl + 1

Lub przejście tam i z powrotem pomiędzy edytorem a drzewem plików:
Ctrl + Shift + e




### Nowy folder, nowy plik
Jestem w explorerze (Ctrl + Shift + e)

{
  "key": "alt+n",
  "command": "explorer.newFile",
  "when": "explorerViewletVisible && filesExplorerFocus && !explorerResourceReadOnly"
},
{
  "key": "alt+shift+n",
  "command": "explorer.newFolder",
  "when": "explorerViewletVisible && filesExplorerFocus && !explorerResourceReadOnly"
}


# TODO - reveal file in explorer
Na Macu nie działa zaden skrot klawiszowy, ktory pozwalalby podswietlic aktualny plik w explorezre. Sprawdzic na linux/windows