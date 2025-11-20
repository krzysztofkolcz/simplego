# Tworzenie struktury go:
wpisuje np.:
```
tenant := model.Tenant{}
```
Kursor ustawiam pomiędzy {}
Windos/Linux: Ctrl + .  -> wybieram z menu 'Fill Struct' (niby plugin gopls)
Mac: command + .

# Wyszukiwanie Ctrl + Shift + f

Windows/Linux: Ctrl + Shift + F

macOS: Cmd + Shift + F

 2. Poruszanie się po wynikach:
Czynność	                                                Skrót
Skok do następnego wyniku	                                F4
Skok do poprzedniego wyniku	                                Shift + F4
Wejście w konkretny wynik (otwarcie pliku na danej linii)	Enter
Zamknięcie panelu wyszukiwania	                            Esc
Powrót do panelu wyszukiwania (z edytora)	                Ctrl + Shift + F (ponownie)

Również strzałki poruszają się po wynikach, jeżeli jestem na liśce wyników.
Tab przechodzi po kolejnych opcjach interfejsu (np. zmiana wyszukiwania na regex.)

# Błędy
Panel 'Problems'
Windows/Linux: Ctrl + Shift + m
macOS: Cmd + Shift + m

błędy (Error),
ostrzeżenia (Warning).

F8 – następny błąd
Shift + F8 – poprzedni błąd

Ctrl + J - ukrywa dolny panel.
Ctrl + B - ukrywa boczny panel.

# Run and Debug
Wygląda na to, ze VS Code nie ma domyslnie Run | Debug dla funkcji main() w main.go.
Moge dodawac konfiguracje w panelu 'Run and Debug' (kolko zebate - launch.json)
Cmd + Shift + D
Ctrl + Shift + D

Zainstalowalem plugin Code Runner.

Plugin golang.go powinien teoretycznie udostepniac 'Run | Debug', ale robi to tylko dla testow

Cmd + Shift + p
Ctrl + Shift + p
Go: Install/Update tools

# Porównanie plików
1. Użycie menu kontekstowego w Explorerze

Otwórz Explorer (Ctrl+Shift+E lub ikona plików po lewej).
Kliknij prawym przyciskiem na pierwszy plik, wybierz Select for Compare.
Kliknij prawym przyciskiem na drugi plik, wybierz Compare with Selected.
VS Code otworzy widok porównania w dwóch kolumnach z podświetleniem różnic.

# todo - 25 VS Code Productivity Tips and Speed Hacks
https://www.youtube.com/watch?v=ifTF3ags0XI


# Nawigacja po elementach pliku
Wyszukiwanie:
Mac: command + f
Windows/linux:


Przechodzenie po elementach pliku z linii komend:
Mac: command + shift + p, @  (najpierw chyba usunąć >)
Windows/linux:

Przechodzenie po elementach w całym projekcie z linii komend:
Mac: command + shift + p, #  (najpierw chyba usunąć >)
Windows/linux:

Przechodzenie po elementach pliku:
Mac: command + shift + .
Windows/linux: 

Przejście do linii nr.:
Mac: control + g

Przenoszenie linii:
Mac: option + up arrow/down arrow
Windows/linux: alt + up arrow/down arrow


Kopiowanie linii:
Mac: option + shift + up arrow/down arrow
Windows/linux: alt + shift + up arrow/down arrow

# Snippet
Jak dodać snippet w VS Code?
## Otwórz edytor snippetów

Wciśnij:

Ctrl + Shift + P → wpisz “Snippets: Configure Snippets” → Enter
→ wybierz go.json (snippet dla Golanga)

To otworzy plik:

%APPDATA%/Code/User/snippets/go.json   (Windows)
~/.config/Code/User/snippets/go.json   (Linux)
~/Library/Application Support/...       (Mac)

## Dodaj własny snippet

Przykład — snippet dla metody z receiverem:
```
{
    "Go receiver method": {
        "prefix": "gorec",
        "body": [
            "func (${1:r} *${2:Type}) ${3:MethodName}(${4:params}) ${5:returnType} {",
            "    ${6:// TODO: implement}",
            "}"
        ],
        "description": "Create a Go method with receiver"
    }
}
```
Co robi ten snippet?

Wpisujesz:

gorec + TAB


I dostajesz gotowy szablon, np.:

func (s *Service) DoSomething(ctx context.Context) error {
    // TODO: implement
}

Pola $1, $2, $3 … są wypełniane kolejno:

$1 – nazwa receivera (np. s)
$2 – typ receivera (Service)
$3 – nazwa metody (DoSomething)
$4 – parametry (ctx context.Context)
$5 – typ zwracany (error)
$6 – treść metody

VS Code pozwala przeskakiwać między nimi Tabem.

## Dodaj więcej snippetów (przykład: metoda bez zwracania)
{
    "Go void method": {
        "prefix": "govoid",
        "body": [
            "func (${1:r} *${2:Type}) ${3:MethodName}(${4:params}) {",
            "    ${5:// TODO: implement}",
            "}"
        ],
        "description": "Go method with receiver and no return"
    }
}


# GIT
## Przejście do panelu Git
ctrl + shift + G, G

## Moje proponowane skroty klawiszowe dla git
Ctrl + g, s -> git stash
Ctrl + g, b -> git checkout to...
Ctrl + g, a -> git stage all changes
Ctrl + g, c -> checkout branch // NIE dziala
Ctrl + g, p -> git pull
Ctrl + g, d -> git branch -d