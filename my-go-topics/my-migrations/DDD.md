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


  