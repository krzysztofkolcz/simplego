https://chatgpt.com/c/690494d1-13fc-8326-8e02-e55b2f6b5371

brew install derailed/k9s/k9s

1) Instalacja i start

Instalacja (przykładowo macOS/Homebrew):
brew install derailed/k9s/k9s

Start: po prostu k9s (użyj KUBECONFIG=... k9s lub k9s -A jeśli chcesz namespace all). 
k9scli.io
+1

2) Podstawowa nawigacja (idea)

: — wejdź w command mode (jak w vimie). Tam wpisujesz typ zasobu, np. :pod (lub skrót :po) aby listę podów w bieżącym namespace. 
Palark
+1

/ — filtr (fuzzy search) w aktualnym widoku.

? — pomoc / pełna lista hotkeyów w danym widoku.

3) Praca na namespace / kontekście

:ns — lista namespace’ów; po wpisaniu możesz przefiltrować (/myns) i Enter, żeby przełączyć.

:ctx — listuje konteksty kubeconfig, możesz się przełączyć. 
k9scli.io

4) Przeglądanie podów i logi (to chyba najważniejsze dla Ciebie)

Wpisz :pod albo :po → wybierz pod (strzałki / wpisz filtr).

Z wybranym padem naciśnij l (mała litera L) lub L — otwiera logi tego poda (log view). Aby wrócić -> Esc. 
KodeKloud Notes
+1

W widoku logów dostępne przydatne skróty (typowe / występują w dokumentacji/tutorialach):

S — zatrzymaj/uruchom autoscroll (stop auto-scrolling).

T — toggle timestamp w logach.

/ — search w treści logów.

p — poprzednie logi dla podu (previous logs) — przydatne gdy kontener został zrestartowany.

0 lub 0 (czasem 0 = tail) — w niektórych wersjach: przełącz na tail/follow (sprawdź ?).

liczby 2, 3 i inne — filtry czasu w niektórych buildach (last minute / last 5 minutes). 
KodeKloud Notes
+1

kubectl odpowiedniki (jeśli potrzebujesz wyjść z k9s albo zrobić z terminala):

kubectl logs -n NAMESPACE POD [-c CONTAINER] -f — tail logs.

kubectl logs -n NAMESPACE POD --previous — poprzednie logi (po restarcie).

5) Exec / shell do kontenera

W k9s: wybierz pod i naciśnij zwykle s lub użyj akcji shell z menu (zależne od wersji/konfigu). Jeśli s nie działa, sprawdź ? w widoku poda — zobaczysz dostępne akcje (exec/shell/describe/edit). (k9s ma też nodeShell feature gate do shellowania węzłów). Jeśli nie chcesz się bawić: kubectl exec -it -n NAMESPACE POD -- /bin/bash lub sh. 
k9scli.io
+1

6) Inne często używane akcje na zasobach

Po zaznaczeniu zasobu (pod/po/dep/..):

d — describe (kube describe).

y — wyświetl YAML.

e — edytuj zasób (otworzy $EDITOR z manifestem).

ctrl-d — usuń (delete) z potwierdzeniem.

ctrl-k — kill (force delete / delete --now) bez potwierdzenia.

shift+f — port‑forward (otwiera UI do wyboru portów i forwardowania). 
k9scli.io
+1

kubectl odpowiedniki: kubectl describe, kubectl apply -f - (dla edycji), kubectl delete pod/..., kubectl port-forward ....

7) Filtry, label selector i wyszukiwanie

W widoku podów wpisz / a następnie np. app=whoami żeby filtrować po labelach. Możesz też wpisać :pod /-l app=whoami w command mode. 
Palark

8) Port‑forward i benchmark

SHIFT+f (port-forward UI) → wybierasz kontener/port → ctrl-b (benchmark) uruchamia prosty load test (zapisuje wyniki do /tmp); to przydatne do szybkich testów. (funkcja bench wymaga konfiguracji). 
hackingnote.com

9) Konfiguracja i hotkeys

Możesz dodać własne hotkeys ($XDG_DATA_HOME/k9s/hotkeys.yaml) i pluginy ($XDG_CONFIG_HOME/k9s/plugins.yaml) — ułatwia przełączanie do często używanych widoków i automatyzację akcji. 
k9scli.io
+1

10) Szybkie przykłady (konkret)

Włącz k9s i przejdź do namespace cmk:
:ns → wpisz cmk → Enter.

Pokaż pody: :pod → wpisz /post aby przefiltrować (np. cmk-post).

Zaznacz pod strzałkami → naciśnij l aby zobaczyć logi → naciśnij S aby zatrzymać autoscroll → / aby wyszukać frazę w logu.

Aby wejść do shella: zaznacz pod → naciśnij s (jeśli działa) — albo otwórz osobny terminal i:
kubectl exec -it -n cmk cmk-postgresql-... -- /bin/bash.

11) Tipy i dobre praktyki

Naciśnij ? często — pokazuje dostępne akcje dla aktualnego widoku (najpewniejsze źródło skrótów dla Twojej wersji). 
Palark

Jeśli robisz ryzykowne działania (np. ctrl-k), pamiętaj, że to force-delete.

Ustaw aliasy / hotkeys do często używanych widoków (np. :po → Twoje pody produkcyjne).

Korzystaj z YAML/edytora (e) do szybkich poprawek (edytuje manifest i robi kubectl apply).

12) Gdzie poczytać więcej (oficjalne i dobre tutoriale)

Dokumentacja / commands & hotkeys (oficjalna): k9scli.io — Commands / Hotkeys. 
k9scli.io
+1

KodeKloud / tutorialy (log view, skróty). 
KodeKloud Notes

GitHub repo k9s (README, issues, przykłady).