# Warstwa ./internal/application
                                                                                                                                                                                              
  Jeden problem: repository/public/tenant_repository.go jest niepełny — ma tylko struct, brak konstruktora i metod. Nie implementuje tenant.Repository.                                       
                                                                                                                                                                                              
  ---                                                                                                                                                                                         
  2. Warstwa Application — gdzie i jak                                                                                                                                                        
                                      
  Tak, internal/application/ to właściwe miejsce. Warstwa aplikacji orchestruje logikę: przyjmuje Command/Query, używa repozytorium (przez interfejs domenowy), zwraca wynik.
                                                                                                                                                                                              
  Dla CQRS typowy podział:                                                                                                                                                                    
                                                                                                                                                                                              
  internal/application/                                                                                                                                                                       
    command/      
      create_todo.go      ← handler dla mutacji
      complete_todo.go                                                                                                                                                                        
      delete_todo.go
    query/                                                                                                                                                                                    
      get_todo.go         ← handler dla odczytów

  Kluczowa zasada: Application layer zna tylko interfejsy domenowe (todo.Repository), nie konkretne implementacje z infrastructure.                                                           
   
  Przykład — command/create_todo.go                                                                                                                                                           
                  
  package command                                                                                                                                                                             
                  
  import (
      "context"
      "github.com/google/uuid"
      "github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
  )                                                                                                                                                                                           
   
  type CreateTodoCommand struct {                                                                                                                                                             
      Title string
  }

  type CreateTodoHandler struct {
      repo todo.Repository
  }

  func NewCreateTodoHandler(repo todo.Repository) *CreateTodoHandler {                                                                                                                        
      return &CreateTodoHandler{repo: repo}
  }                                                                                                                                                                                           
                  
  func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error) {
      t := todo.Todo{
          ID:    uuid.New(),                                                                                                                                                                  
          Title: cmd.Title,
      }                                                                                                                                                                                       
      if err := h.repo.Create(ctx, t); err != nil {
          return uuid.Nil, err
      }
      return t.ID, nil                                                                                                                                                                        
  }
                                                                                                                                                                                              
  Przykład — query/get_todo.go

  package query

  import (
      "context"
      "github.com/google/uuid"
      "github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
  )

  type GetTodoResult struct {
      ID        uuid.UUID
      Title     string
      Completed bool
  }

  type GetTodoHandler struct {
      repo todo.Repository
  }                                                                                                                                                                                           
   
  func NewGetTodoHandler(repo todo.Repository) *GetTodoHandler {                                                                                                                              
      return &GetTodoHandler{repo: repo}
  }

  func (h *GetTodoHandler) Handle(ctx context.Context, id uuid.UUID) (*GetTodoResult, error) {                                                                                                
      t, err := h.repo.GetByID(ctx, id)
      if err != nil {                                                                                                                                                                         
          return nil, err
      }
      return &GetTodoResult{ID: t.ID, Title: t.Title, Completed: t.Completed}, nil
  }                                                                                                                                                                                           
   
  ---                                                                                                                                                                                         
  3. Jak to spinać w cmd/api/main.go
                                                                                                                                                                                              
  // Infrastruktura
  txManager := db.NewTxManager(pool)                                                                                                                                                          
  commandQ := commanddb.New(pool)
  queryQ   := querydb.New(pool)
  todoRepo := tenantrepo.NewTodoRepository(commandQ, queryQ)
                                                                                                                                                                                              
  // Application
  createTodoHandler := command.NewCreateTodoHandler(todoRepo)                                                                                                                                 
  getTodoHandler    := query.NewGetTodoHandler(todoRepo)

  ---                                                                                                                                                                                         
  4. Diagram zależności
                                                                                                                                                                                              
  cmd/api/main.go           ← wiring (zna wszystkie warstwy)
      │                                                                                                                                                                                       
      ▼
  internal/application/     ← orchestracja, zna tylko interfejsy domenowe                                                                                                                     
      command/    
      query/                                                                                                                                                                                  
      │           
      ▼ (przez interfejs)
  internal/domain/          ← czysta logika, zero zależności zewnętrznych
      todo/repository.go    ← interface                                                                                                                                                       
      todo/entity.go
      │                                                                                                                                                                                       
      ▲ (implementuje interfejs)
  internal/infrastructure/  ← szczegóły techniczne                                                                                                                                            
      repository/           ← konkretne implementacje                                                                                                                                         
      db/                   ← sqlc, pool, migrations
                                                                                                                                                                                              
  ---             
  5. Droga do CQRS i Event Sourcing                                                                                                                                                           
                                   
  Masz już dobry fundament — oddzielone sqlc/command i sqlc/query. Naturalne kolejne kroki:
                                                                                                                                                                                              
  Krok 1 (teraz): Application layer jak powyżej — command/query handlers                                                                                                                      
  Krok 2 (CQRS pełne): Osobna baza danych / osobna projekcja dla read modelu                                                                                                                  
  Krok 3 (Event Sourcing): Command handler emituje DomainEvent zamiast bezpośrednio pisać do repo — zdarzenia są źródłem prawdy, read model jest projekcją zdarzeń                            
                                                                                                                                                                                              
  Na razie najważniejsze — zbuduj Application layer. Chcesz, żebym uzupełnił TenantRepository i dodał pełny przykład command/create_todo.go do kodu? 


  ❯ W którym miejscu używać WithinTransactionReadonly i WithinTransaction, żeby mieć transakcje?                                                                                                
                                                                                                                                                                                              
● To jest kluczowe pytanie w DDD. Problem w obecnym kodzie polega na tym, że transakcje są zarządzane przez test (wywołującego), a nie przez command handler. W produkcyjnej aplikacji to jest
   złe miejsce.                                                                                                                                                                               

  ---                                                                                                                                                                                         
  Gdzie powinny być transakcje?                                                                                                                                                               
                                                                                                                                                                                              
  Zasada: granica transakcji = granica use case'a = command handler.                                                                                                                          
                                                                                                                                                                                              
  Każdy command handler powinien sam otwierać i commitować transakcję. Ale command handler nie powinien znać infrastruktury (TxManager, commanddb.Queries) — to naruszałoby DDD.              
                                                                                                                                                                                              
  Rozwiązanie: Unit of Work — interfejs domenowy, implementacja w infrastrukturze.                                                                                                            
                                                                                                                                                                                              
  ---                                                                                                                                                                                         
  Unit of Work w tym projekcie
                              
  1. Interfejs w internal/domain/todo/ (czysta domena, zero importów infrastruktury):
                                                                                                                                                                                              
  // internal/domain/todo/unit_of_work.go
  package todo                                                                                                                                                                                
                  
  import "context"                                                                                                                                                                            
                  
  type UnitOfWork interface {
      Execute(ctx context.Context, fn func(repo Repository) error) error
  }

  2. Implementacja w internal/infrastructure/repository/tenant/:                                                                                                                              
   
  // internal/infrastructure/repository/tenant/unit_of_work.go                                                                                                                                
  package repository

  import (
      "context"
      "github.com/krzysztofkolcz/mymigrations/internal/domain/todo"                                                                                                                           
      "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"                                                                                                                     
      querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"                                                                                                  
      commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"                                                                                              
  )                                                                                                                                                                                           
                                                                                                                                                                                              
  type TodoUnitOfWork struct {                                                                                                                                                                
      txManager    *db.TxManager
      tenantSchema string
  }

  func NewTodoUnitOfWork(txManager *db.TxManager, tenantSchema string) *TodoUnitOfWork {                                                                                                      
      return &TodoUnitOfWork{txManager: txManager, tenantSchema: tenantSchema}
  }                                                                                                                                                                                           
                  
  func (u *TodoUnitOfWork) Execute(ctx context.Context, fn func(todo.Repository) error) error {                                                                                               
      return u.txManager.WithinTransaction(ctx, u.tenantSchema, func(commandQ *commanddb.Queries) error {
          queryQ := querydb.New(commandQ.DB())                                                                                                                                                
          repo := NewTodoRepository(commandQ, queryQ)                                                                                                                                         
          return fn(repo)
      })                                                                                                                                                                                      
  }               

  3. Command handler używa tylko interfejsów domenowych:                                                                                                                                      
   
  // internal/application/command/create_todo.go                                                                                                                                              
  type CreateTodoHandler struct {
      uow todo.UnitOfWork
  }                                                                                                                                                                                           
   
  func NewCreateTodoHandler(uow todo.UnitOfWork) *CreateTodoHandler {                                                                                                                         
      return &CreateTodoHandler{uow: uow}
  }

  func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error) {                                                                                         
      t := todo.Todo{ID: uuid.New(), Title: cmd.Title}
                                                                                                                                                                                              
      err := h.uow.Execute(ctx, func(repo todo.Repository) error {
          return repo.Create(ctx, t)                                                                                                                                                          
      })          
      if err != nil {
          return uuid.Nil, err
      }
      return t.ID, nil                                                                                                                                                                        
  }
                                                                                                                                                                                              
  4. Queries używają WithinTransactionReadonly — ale gdzie?                                                                                                                                   
   
  Query handlery nie mutują stanu, więc nie potrzebują Unit of Work. Mogą dostać todo.Repository bezpośrednio, skonfigurowane z readonly tx — albo mieć własny interfejs ReadRepository:      
                  
  // internal/application/query/get_todo.go                                                                                                                                                   
  type GetTodoHandler struct {                                                                                                                                                                
      repo todo.Repository  // skonfigurowane z WithinTransactionReadonly
  }                                                                                                                                                                                           
                  
  W HTTP handlerze lub w miejscu wiring (main.go):                                                                                                                                            
  // Wiring — cmd/api/main.go lub app.go
  tenantSchema := "tenant_xyz"  // z JWT/kontekstu requestu                                                                                                                                   
                                                                                                                                                                                              
  // Commands — przez Unit of Work (zarządza transakcją wewnętrznie)                                                                                                                          
  uow := tenantrepo.NewTodoUnitOfWork(txManager, tenantSchema)                                                                                                                                
  createHandler := command.NewCreateTodoHandler(uow)                                                                                                                                          
                                                                                                                                                                                              
  // Queries — repo z readonly tx (lub pool bez tx dla prostych przypadków)
  queryRepo := tenantrepo.NewTodoRepository(nil, querydb.New(pool))                                                                                                                           
  getHandler := query.NewGetTodoHandler(queryRepo)                 
                                                                                                                                                                                              
  ---             
  Podsumowanie — gdzie co żyje                                                                                                                                                                
   
  tests/integration/     → WithinTransaction / WithinTransactionReadonly                                                                                                                      
                            (bo test zastępuje HTTP handler w roli wywołującego)                                                                                                              
                                                                                                                                                                                              
  cmd/api / HTTP handler → tworzy UnitOfWork z tenantSchema z kontekstu requestu                                                                                                              
                            NIGDY nie zarządza transakcją bezpośrednio                                                                                                                        
                                                                                                                                                                                              
  application/command/   → wywołuje uow.Execute() — nie wie o tx                                                                                                                              
  application/query/     → używa repo bezpośrednio (readonly) — nie wie o tx
                                                                                                                                                                                              
  infrastructure/        → WithinTransaction i WithinTransactionReadonly                                                                                                                      
                            tylko tutaj, ukryte za interfejsami         