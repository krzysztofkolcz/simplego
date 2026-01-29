# 1. Podstawowe podejrzenie rejestru w trybie normalnym

Wpisz w trybie normalnym (czyli Esc):

:registers


lub krócej:

:reg

# 2. Kopiowanie do systemowego schowka

Rejestr "+ = schowek systemowy (Ctrl+C / Ctrl+V)

Rejestr "* = schowek „primary” (zaznaczenie myszką w systemach Linux)

Możesz podejrzeć tylko te dwa rejestry, np.:

:reg +
:reg *
# 5. Bonus: szybki dostęp do schowka systemowego

Jeśli chcesz zsynchronizować Vima z systemowym schowkiem na stałe:

"vim.useSystemClipboard": true


Dzięki temu wszystko, co kopiujesz w Vimie (yy, p itd.), trafia też do systemowego schowka i można podejrzeć przez Ctrl+V lub :reg +.


# Wyłączenie przechwytywania Ctrl przez plugin vim
Settings: 
MAC : Command + ,
vim: use ctrl keys - wyłączyć