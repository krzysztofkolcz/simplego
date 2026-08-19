Skąd się biorą dane o wycieku - przypuszczam, że parsowanie stron dark net? Może jakieś raporty firm o atakach zaweirające lista adresów email?
Dopasowanie wycieku do usera - wyszukiwanie po emailu/nr telefonu/peselu. Założony indeks na te pola. Zapis klucza do tabeli w bazie danych.

Czyli
Pobranie emaila z wycieku -> rządanie API sprawdzenia emaila, czy mamy w bazie (lub batch - kilkaset maili) -> wrzucenie do kolejki taska (Rabbit MQ) -> pobranie przez worker -> sprawdzenie w bazie (indeks) -> zapis do tabeli raportu.

Jezeli przetworzenie eventu się nie uda -> retry.

PII - nie wiem co to

50 mln rekorwów -> kubernetes, uruchamianie dodatkowych workerów (ale nie wiem jak)