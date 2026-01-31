Czy dobrze rozumiem:

WORKDIR /app
 ustawia folder w kontenerze na /app

COPY go.mod go.sum ./
 kopiuje go.mod i go.sum do katalogu /app na kontenterze

RUN go mod download
ściąga biblioteki do kontenera (tylko te, których poprzednio nie miał?)

COPY . .
 kopiuje całość kodu do kontenera oprócz .dockerignore

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \ go build -o app ./cmd/server
buduje aplikację do binarki /app/app

FROM gcr.io/distroless/base-debian12
obraz, na jakim zostanie uruchomiona binarka

WORKDIR /app
 ustawia workdir na /app

COPY --from=builder /app/app /app/app
 nie wiem, co to robi? Kopiuje z poprzedniego kroku do aktualnego? Jak to rozumieć? Że krok budowania ma inny system plikow niż krok runnera?

EXPOSE 8080
 otwiera port 8080 kontenera?

USER nonroot:nonroot
 użytkownik w kontenerze

ENTRYPOINT ["/app/app"]
 uruchomienie binarki?



Tak — rozumiesz to w ~95% idealnie 👌
Doprecyzujmy te 5%, bo tam kryje się senior-level understanding Dockera + Go.

Przejdę linia po linii, potwierdzając lub korygując.

1️⃣ WORKDIR /app

ustawia folder w kontenerze na /app

✅ Dokładnie tak

Jeśli katalog nie istnieje → Docker go tworzy

Każde kolejne COPY, RUN, CMD działa względem /app

To jest jak:

cd /app

2️⃣ COPY go.mod go.sum ./

kopiuje go.mod i go.sum do /app

✅ Tak
Finalnie masz:

/app/go.mod
/app/go.sum


🎯 To jest klucz do cache Dockera (o tym za chwilę).

3️⃣ RUN go mod download

ściąga biblioteki do kontenera (tylko te, których poprzednio nie miał?)

✅ Tak — i to bardzo ważne jak Docker to robi

Go pobiera zależności do:

/go/pkg/mod


Docker cache’uje ten layer

➡️ Jeśli go.mod / go.sum się nie zmienią:

ten krok NIE wykona się ponownie

build jest bardzo szybki

To jest celowy trik architektoniczny.

4️⃣ COPY . .

kopiuje całość kodu do kontenera oprócz .dockerignore

✅ Tak

Czyli:

cmd/

internal/

Makefile

Helm chart

itd.

➡️ Ten krok psuje cache, ale nie psuje cache depsów
➡️ Dokładnie o to chodzi

5️⃣ RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/server

buduje aplikację do binarki /app/app

✅ W punkt

./cmd/server → katalog z main.go

-o app → wynik to:

/app/app


CGO_ENABLED=0 → statyczna binarka

GOOS/GOARCH → target Linux AMD64

🧠 Efekt końcowy:

/app/app   ← jeden plik, cała aplikacja

🔥 TERAZ NAJWAŻNIEJSZA CZĘŚĆ
6️⃣ FROM gcr.io/distroless/base-debian12

obraz, na jakim zostanie uruchomiona binarka

✅ Tak
ALE:

🧠 To jest NOWY OBRAZ, NOWY SYSTEM PLIKÓW

nie ma Go

nie ma /app

nie ma nic z poprzedniego stage’a

to inny świat

7️⃣ COPY --from=builder /app/app /app/app

nie wiem, co to robi?
Kopiuje z poprzedniego kroku do aktualnego?

🎯 TAK. I to jest sedno multi-stage buildów.

Jak to czytać po ludzku:

„Skopiuj plik /app/app
z obrazu o nazwie builder
do /app/app w AKTUALNYM obrazie”

Co to oznacza w praktyce?

Masz DWA RÓŻNE SYSTEMY PLIKÓW:

🔵 Stage 1 – builder
/app
├── app        ← binarka
├── go.mod
├── go.sum
├── cmd/
├── internal/
└── /go/pkg/mod

🟢 Stage 2 – runtime (distroless)
/app
└── app        ← TYLKO binarka


➡️ NIC WIĘCEJ NIE PRZECHODZI

To dlatego:

obraz jest mały

nie ma source code

nie ma Go

nie ma sekretów

nie ma attack surface

🔥 To jest production best practice

8️⃣ EXPOSE 8080

otwiera port 8080 kontenera?

⚠️ Prawie, ale nie do końca

❗ EXPOSE:

NICZEGO nie otwiera

to tylko metadana / dokumentacja

To mówi:

„Ta aplikacja oczekuje, że będzie nasłuchiwać na 8080”

Port faktycznie mapujesz przy:

docker run -p 8080:8080

9️⃣ USER nonroot:nonroot

użytkownik w kontenerze

✅ Tak

distroless ma już użytkownika nonroot

aplikacja nie działa jako root

ogromny plus bezpieczeństwa (CVE, K8s, audit)

🔟 ENTRYPOINT ["/app/app"]

uruchomienie binarki?

✅ Tak — dokładnie

Po starcie kontenera Docker robi:

/app/app


Jeśli dodasz argumenty:

docker run myapp --port=8080


To będzie:

/app/app --port=8080

🧠 TL;DR — rozumiesz to bardzo dobrze

Twoje rozumienie jest:

✅ technicznie poprawne

✅ zgodne z best practices

✅ gotowe na Kubernetes / CI

Jedyna korekta:

EXPOSE nie „otwiera” portu — tylko go opisuje

COPY --from=builder = kopiowanie między DWOMA OBRAZAMI