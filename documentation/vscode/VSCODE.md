# Tworzenie struktury go:
wpisuje np.:
```
tenant := model.Tenant{}
```
Kursor ustawiam pomiędzy {}
Ctrl + .  -> wybieram z menu 'Fill Struct' (niby plugin gopls)

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
