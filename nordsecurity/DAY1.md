1. 
Rozumiem, że to struktura, otrzymana jako źródło wycieku. Jeżeli tak:

type LeakRecord struct {
    Email    string // hashuje - do porównania z emailem w bazie
    Phone    string // hashuje
    Password string // hashuje
    Source   string // rozumiem, że to źródło wycieku? pozostawiam plaintext
}
Zakładam, że w bazie trzymamy te dane zahashowane, więc w celu porównania z danymi z bazie, musze je zahashować.

2. 
Key management - trzymałbym w KMS (sam pracuje przy KMS2.0 SAP). Albo np. w AWS.
Rotacja - nie mam pomysłu. Jeżeli trzymamy tylko hashe wartości (np. emaila, hasła, peselu, danych finansowych), to nie jesteśmy w stanie odzyskać oryginalnej wartości. Jeżeli trzymamy zakodowane (encoded), to dekodujemy i kodujemy ponownie nowym kluczem. Operacja ta musiała by lecieć w batchu.

3. 
DPO - user żąda usunięcia danych
Oznaczyłbym dane usera 'do usunięcia' (nazwa statusu np. TO_DELETE).
Uruchamiam serwis usunięcia
Usunięcie z bazy, oraz propagacja do replik.
Backupu nie ruszam
Serwisy - nie do końca rozumiem, chodzi o trzymanie jego danych w pamięci aplikacji?
Kolejki - jest tutaj pewien okres, od uzyskania statusu TO_DELETE nie wysyłam już jego danych do kolejek. Niestety nie wiem, czy jest jakiś znacznik, który moge wysłać. Jeżeli tak, to byłoby dobre. Wszystkie kolejki po tym znaczniku nie miałyby juz danych do usunięcia.
Na koniec po otrzymaniu znacznika, rekord z danymi użytkownika jest usuwany i propaguje się do replik.
Backapów nie ruszam.