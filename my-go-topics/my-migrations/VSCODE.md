# VS Code - snippety
❯ Jak otworzyć snippety kodu?                                                                                                                                           

W VS Code są trzy sposoby:

  1. Command Palette (najszybciej)
  Ctrl+Shift+P → wpisz snippets → "Preferences: Configure User Snippets"

  2. Menu
  File → Preferences → Configure User Snippets

  3. Bezpośrednio przez plik (już otwarty w IDE)
  ~/.config/Code/User/snippets/globalsnippet.code-snippets — to globalny plik dla wszystkich języków.

  ---
  Dla snippetów tylko dla Go wybierz "Go" po otwarciu przez Command Palette — VS Code zapisze do ~/.config/Code/User/snippets/go.json. Snippety działają w obu plikach,
  różnica tylko w zakresie języka.

# go tools
go install golang.org/x/tools/cmd/impl@latest
impl 'r *ComponentRepository' catalog.ComponentRepository

  # 1. Piszesz w internal/domain/catalog/repository.go:
  type ComponentRepository interface {
      Create(ctx context.Context, c Component) error
      GetByID(ctx context.Context, id uuid.UUID) (*Component, error)
      List(ctx context.Context) ([]Component, error)
  }

# 2. W katalogu repository/tenant/catalog/ wywołujesz:
```
impl 'r *ComponentRepository' github.com/krzysztofkolcz/mymigrations/internal/domain/catalog.ComponentRepository

```
  impl generuje wszystkie sygnatury z panic("not implemented") — nie musisz ręcznie przepisywać nazw metod i typów parametrów.

 VS Code Go extension ma to wbudowane — kursor na nazwę struktury → Ctrl+. → "Implement interface" → wpisujesz nazwę interfejsu.
2. VS Code snippets — dla powtarzalnych wzorców
Plik ~/.config/Code/User/snippets/go.json (lub File → Preferences → User Snippets → Go):