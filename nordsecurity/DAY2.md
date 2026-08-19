1. idempotency key - hash(batch_id + email) lub hash(batch_id + pesel) - w zależności od typu pliku importu. ON CONFLICT DO NOTHING.

2. Konkretne błędy w workerze matchującym, np.: 
błąd połączenia z bazą danych - transient - czyli muszę naprawić połączenie, nie jest to błąd danych.
rekord z pliku ma brakujące/zepsute pole - permanent - to problem danych dla tego batcha. Jeżeli dostanę plik z poprawnymi danymi, to będzie już kolejny batch
naruszenie unique constraint - jeżeli uniuqe nie jest na idempotency key, to jest jakiś błąd w danych. czyli dwa razy pojawia się email dla danego batcha. Czyli permanent - bo błędny plik wsadowy.
worker nie potrafi zdeserializować wiadomości - permanent - błąd w kodzie, lub pliku wsadowym
zewnętrzny serwis odpowiada timeout - transient. Można robić retry.
unique constraint validation na idempotency_key - już przetworzone. ACK

3. circuit braker - jeżeli chodzi o połączenie z zewnętrznym api, no to oznaczenie, żeby się nie dobijać, dopóki serwis nie dizała. Co jakiś czas sprawdzenie czy juz działa. Jak chodzi o serwis wewnętrzny, no to podobnie, zeby nie marnować zasobów. Kubernetes powinien zrestartować serwis, a jeżeli błąd będzie się powtarzał, to potrzebna interwencja