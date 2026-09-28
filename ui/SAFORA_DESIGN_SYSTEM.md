# Safora — Design System

> **Open-source automated backup management.**

Documento de referência visual e de interface do Safora.  
O objetivo é manter uma identidade consistente entre dashboard, documentação, landing page, CLI, ícones, notificações e demais interfaces do projeto.

---

## 1. Identidade

### Nome

**Safora**

Sempre utilizar `Safora` com S maiúsculo e restante em minúsculas.

Evitar:
- SAFORA
- SAFORA Backup
- Safora Backup Manager
- Safora Cloud

O nome deve permanecer independente do meio de armazenamento. O produto gerencia backups para discos locais, NAS, servidores e serviços de armazenamento em nuvem.

### Posicionamento

Safora é uma ferramenta open source de infraestrutura para automatizar e centralizar o gerenciamento de backups.

A identidade deve comunicar:

- segurança
- confiabilidade
- automação
- infraestrutura
- simplicidade
- tecnologia
- transparência/open source

A marca não deve parecer um antivírus, empresa de segurança física ou serviço genérico de armazenamento em nuvem.

---

# 2. Logo

## 2.1 Conceito

O símbolo do Safora é um **S abstrato construído como uma fita/ribbon geométrica**.

O símbolo representa:

- o "S" de Safora;
- continuidade;
- fluxo de dados;
- ciclo de backup;
- movimentação entre origem e destino;
- proteção de dados.

A construção geométrica e arredondada evita uma aparência excessivamente corporativa ou militar.

## 2.2 Estrutura

O logo possui dois elementos:

1. Símbolo geométrico "S";
2. Wordmark `Safora`.

O símbolo pode ser utilizado isoladamente como ícone de aplicação, favicon ou identificação compacta.

## 2.3 Versões

### Logo principal

Símbolo à esquerda + `Safora` à direita.

Uso preferencial em:
- website;
- documentação;
- README;
- dashboard;
- apresentações;
- materiais institucionais.

### Símbolo

Somente o símbolo "S".

Uso em:
- favicon;
- app icon;
- avatar;
- sidebar compacta;
- CLI;
- notificações;
- pequenos espaços.

### Versão para fundo escuro

Símbolo em gradiente teal/ciano + wordmark claro.

### Versão para fundo claro

Símbolo em gradiente teal/ciano + wordmark em navy.

### Versão monocromática

Quando o uso de cores não for possível, utilizar o símbolo e o wordmark em uma única cor.

---

# 3. Área de proteção

A área livre ao redor do logo deve ser equivalente, no mínimo, à largura de uma das barras principais do símbolo.

Nunca posicionar textos, bordas, ícones ou outros elementos dentro dessa área.

Não aplicar:
- sombras pesadas;
- contornos;
- distorções;
- rotação;
- perspectiva;
- efeitos 3D;
- gradientes diferentes do padrão da marca.

---

# 4. Paleta de cores

## 4.1 Cores principais

| Token | Hex | Uso |
|---|---|---|
| `safora-teal-500` | `#00D1B2` | Cor principal da marca |
| `safora-teal-600` | `#00BFA6` | Ações e estados ativos |
| `safora-teal-700` | `#009B88` | Hover e elementos de maior contraste |
| `safora-cyan-400` | `#19E6D0` | Destaques e gradientes |
| `safora-navy-950` | `#0B1220` | Fundo principal escuro |
| `safora-navy-900` | `#0F172A` | Superfícies escuras |
| `safora-navy-800` | `#172033` | Cards e elementos elevados |
| `safora-white` | `#FFFFFF` | Texto em fundo escuro |

## 4.2 Neutros

| Token | Hex | Uso |
|---|---|---|
| `gray-50` | `#F8FAFC` | Background claro |
| `gray-100` | `#F1F5F9` | Superfícies |
| `gray-200` | `#E2E8F0` | Bordas |
| `gray-300` | `#CBD5E1` | Elementos secundários |
| `gray-400` | `#94A3B8` | Texto auxiliar |
| `gray-500` | `#64748B` | Texto secundário |
| `gray-600` | `#475569` | Texto intermediário |
| `gray-700` | `#334155` | Texto forte |
| `gray-800` | `#1E293B` | Texto principal |
| `gray-900` | `#0F172A` | Texto máximo |

## 4.3 Estados

| Estado | Cor | Hex |
|---|---|---|
| Success | Green | `#22C55E` |
| Warning | Amber | `#F59E0B` |
| Error | Red | `#EF4444` |
| Info | Blue | `#3B82F6` |
| Running | Teal | `#00BFA6` |
| Disabled | Gray | `#94A3B8` |

Os estados devem ser acompanhados por texto ou ícone quando a informação for importante. Não depender exclusivamente de cor.

---

# 5. Gradiente da marca

O gradiente é reservado principalmente para o símbolo e elementos de destaque.

```css
linear-gradient(135deg, #19E6D0 0%, #00D1B2 50%, #009B88 100%)
```

Uso recomendado:
- logo;
- ícone principal;
- pequenos highlights;
- gráficos ou elementos decorativos específicos.

Evitar usar o gradiente em:
- grandes áreas de background;
- todos os botões;
- todos os cards;
- textos extensos.

O gradiente deve funcionar como assinatura visual, não como decoração dominante.

---

# 6. Tipografia

## Família principal

**Outfit**

A tipografia deve ser moderna, geométrica e amigável, equilibrando tecnologia com acessibilidade.

Pesos utilizados:

- 400 — Regular
- 500 — Medium
- 600 — SemiBold
- 700 — Bold

## Família monoespaçada

Para elementos técnicos:

**JetBrains Mono**

Uso em:
- terminal;
- comandos;
- paths;
- IDs;
- hashes;
- logs;
- nomes de arquivos;
- valores técnicos;
- código.

---

# 7. Escala tipográfica

| Token | Tamanho | Peso | Uso |
|---|---:|---:|---|
| `display` | 48px | 700 | Hero / páginas institucionais |
| `h1` | 32px | 600 | Título principal |
| `h2` | 24px | 600 | Seções |
| `h3` | 20px | 600 | Subseções |
| `h4` | 16px | 600 | Títulos de cards |
| `body-lg` | 16px | 400 | Texto destacado |
| `body` | 14px | 400 | Texto padrão |
| `body-sm` | 13px | 400 | Texto secundário |
| `caption` | 12px | 500 | Metadados |
| `mono` | 13px | 400 | Dados técnicos |

Line-height recomendado:

- títulos: `1.15–1.25`
- corpo: `1.5`
- elementos compactos: `1.3`

---

# 8. Espaçamento

Utilizar uma escala baseada em múltiplos de 4px.

| Token | Valor |
|---|---:|
| `space-1` | 4px |
| `space-2` | 8px |
| `space-3` | 12px |
| `space-4` | 16px |
| `space-5` | 20px |
| `space-6` | 24px |
| `space-8` | 32px |
| `space-10` | 40px |
| `space-12` | 48px |
| `space-16` | 64px |
| `space-20` | 80px |
| `space-24` | 96px |

Priorizar consistência em vez de valores arbitrários.

---

# 9. Border radius

O Safora utiliza cantos suavemente arredondados.

| Token | Valor |
|---|---:|
| `radius-sm` | 6px |
| `radius-md` | 8px |
| `radius-lg` | 12px |
| `radius-xl` | 16px |
| `radius-2xl` | 20px |
| `radius-full` | 9999px |

Uso:

- inputs: `8px`
- buttons: `8px`
- cards: `12px`
- modais: `16px`
- badges: `9999px`

Evitar excesso de arredondamento. A interface deve continuar técnica e profissional.

---

# 10. Sombras

As sombras devem ser discretas.

```css
/* Small */
0 1px 2px rgba(15, 23, 42, 0.05);

/* Medium */
0 4px 12px rgba(15, 23, 42, 0.08);

/* Large */
0 12px 32px rgba(15, 23, 42, 0.12);
```

Em dark mode, priorizar contraste por meio de bordas e diferença de superfície em vez de sombras fortes.

---

# 11. Dark Mode

O dark mode é uma parte importante da identidade do Safora, especialmente por ser uma ferramenta de infraestrutura.

Background:

```text
#0B1220
```

Surface:

```text
#0F172A
```

Elevated:

```text
#172033
```

Border:

```text
#263247
```

Texto principal:

```text
#F8FAFC
```

Texto secundário:

```text
#94A3B8
```

Accent:

```text
#00D1B2
```

O teal deve ser usado para indicar interação, estado ativo e informação relevante.

---

# 12. Componentes

## 12.1 Button

### Primary

Background:

`#00BFA6`

Texto:

`#FFFFFF`

Hover:

`#009B88`

### Secondary

Background:

`#172033`

Border:

`#263247`

Texto:

`#E2E8F0`

### Ghost

Background:

transparent

Texto:

`#94A3B8`

Hover:

`#172033`

### Destructive

Background:

`#EF4444`

Usar somente para ações potencialmente destrutivas.

Exemplo:

`Excluir backup`

Nunca utilizar vermelho para ações comuns.

---

# 13. Cards

Cards devem ser utilizados para agrupar informações relacionadas.

Dark:

```text
background: #0F172A
border: #263247
radius: 12px
```

Light:

```text
background: #FFFFFF
border: #E2E8F0
radius: 12px
```

Cards não devem parecer excessivamente elevados.

Priorizar hierarquia por:
1. espaçamento;
2. tipografia;
3. contraste;
4. borda;
5. sombra.

---

# 14. Status de backup

O status é um dos elementos mais importantes da interface.

### Sucesso

Ícone de check.

```text
Backup concluído
Última execução: hoje, 03:00
```

Cor: `#22C55E`

### Em execução

Ícone de loading/progresso.

```text
Backup em execução
42% concluído
```

Cor: `#00BFA6`

### Atenção

```text
Backup concluído com avisos
```

Cor: `#F59E0B`

### Falha

```text
Backup falhou
```

Cor: `#EF4444`

### Nunca executado

```text
Nenhum backup executado
```

Cor: `#94A3B8`

---

# 15. Ícones

Biblioteca recomendada:

**Lucide Icons**

Características:

- outline;
- geometria simples;
- stroke consistente;
- cantos suaves.

Evitar misturar bibliotecas de ícones diferentes.

Tamanhos padrão:

- 14px — elementos compactos;
- 16px — interface;
- 20px — navegação;
- 24px — destaque;
- 32px+ — empty states ou hero.

---

# 16. Linguagem visual

A interface do Safora deve parecer:

**Técnica + simples + confiável + moderna.**

Não deve parecer:

- software bancário;
- painel de monitoramento excessivamente complexo;
- ferramenta militar;
- antivírus;
- dashboard cheio de gráficos sem utilidade.

O usuário deve conseguir responder rapidamente:

1. Meus backups estão funcionando?
2. Qual foi o último backup?
3. Para onde os dados estão sendo enviados?
4. Existe algum problema?
5. Quanto espaço estou utilizando?
6. Quando será o próximo backup?

---

# 17. Dashboard

O dashboard deve priorizar estado operacional.

Hierarquia recomendada:

### Resumo

- Backups funcionando
- Backups com problemas
- Destinos ativos
- Espaço utilizado

### Atividade recente

Lista dos últimos backups:

```text
Servidor Produção
✓ Concluído
Hoje, 03:00
128 GB

Servidor Banco
✓ Concluído
Hoje, 02:30
42 GB

Workstation Dev
⚠ Avisos
Ontem, 22:15
18 GB
```

### Próximas execuções

Mostrar os próximos jobs agendados.

---

# 18. Navegação

A navegação principal deve ser simples.

Estrutura sugerida:

```text
Dashboard

Backups
  Jobs
  Histórico

Destinos

Dispositivos

Armazenamento

Alertas

Configurações
```

Evitar mais de dois níveis de navegação sempre que possível.

---

# 19. Empty states

Empty states devem explicar o próximo passo.

Evitar:

```text
Nenhum dado.
```

Preferir:

```text
Nenhum backup configurado

Configure seu primeiro job de backup para começar a proteger seus dados.

[ Criar backup ]
```

---

# 20. Alertas e notificações

As notificações devem ser objetivas.

Exemplo:

```text
Backup concluído

Servidor Produção foi sincronizado com sucesso.
128 GB enviados para NAS Principal.

03:02
```

Em caso de erro:

```text
Backup falhou

O backup de Servidor Banco não foi concluído.

Motivo: destino indisponível.

[ Ver detalhes ]
```

Nunca esconder o motivo técnico do erro quando ele estiver disponível.

---

# 21. Dados técnicos

Informações técnicas devem utilizar JetBrains Mono quando isso melhorar a leitura.

Exemplos:

```text
/home/data
s3://safora-production/backups
SHA256: 4a9f...
backup-job-01
```

Paths, hashes, IDs e comandos devem ser visualmente diferenciados do conteúdo comum.

---

# 22. Terminal / CLI

A CLI faz parte da identidade do Safora.

Exemplo:

```text
$ safora backup run production

Safora
────────────────────────────────

Source      /var/lib/application
Destination s3://backups/production

Uploading...
████████████████████░░░░ 82%

✓ Backup completed successfully

Duration    02m 41s
Transferred 4.82 GB
```

A saída deve ser limpa e legível.

Não utilizar excesso de ASCII art.

---

# 23. Ilustrações

Quando forem necessárias ilustrações, utilizar:

- formas geométricas;
- linhas simples;
- elementos de infraestrutura;
- servidores;
- armazenamento;
- fluxos de dados;
- nuvem;
- discos;
- conexões.

Evitar:
- cadeados gigantes;
- escudos literais em excesso;
- hackers;
- personagens;
- imagens genéricas de data center;
- estética cyberpunk.

O conceito de proteção deve ser comunicado principalmente pelo sistema visual, não por clichês.

---

# 24. Fotografia

Quando fotografias forem utilizadas, preferir:

- infraestrutura real;
- servidores;
- racks;
- armazenamento;
- workstations;
- profissionais de tecnologia;
- ambientes técnicos.

Tratamento:

- escuro;
- alto contraste;
- pouco saturado;
- teal como destaque.

A fotografia nunca deve competir com o logo.

---

# 25. Landing page

A landing page deve seguir uma estética:

**Dark technical + clean SaaS.**

Hero:

```text
Automated backups.
Without the complexity.

Safora is an open-source platform for managing
automated backups across servers and computers.

[ Get started ] [ View on GitHub ]
```

O hero deve priorizar produto e conceito.

Evitar hero excessivamente decorativo.

---

# 26. Open Source

O fato de o Safora ser open source é parte importante da identidade.

A comunicação visual pode utilizar:

- GitHub;
- terminal;
- documentação;
- arquitetura;
- código;
- comunidade;
- transparência.

A marca não deve parecer uma empresa SaaS fechada.

Elementos como:

```text
Open Source
Self-hosted
Automated
Reliable
```

podem aparecer como atributos do produto.

---

# 27. Acessibilidade

Contraste mínimo deve ser considerado em todos os componentes.

Nunca utilizar somente cor para representar:

- sucesso;
- erro;
- alerta;
- status.

Sempre que possível utilizar:

- ícone;
- texto;
- cor.

Elementos interativos devem possuir estados claros:

- default;
- hover;
- focus;
- active;
- disabled;
- loading;
- error.

---

# 28. Princípios de UI

### 1. Status primeiro

O usuário precisa saber rapidamente se seus dados estão protegidos.

### 2. Complexidade sob demanda

Informações avançadas devem existir, mas não dominar a interface.

### 3. Automação visível

O sistema deve mostrar o que está acontecendo automaticamente.

### 4. Erros explicáveis

Mensagens de erro devem ajudar o usuário a resolver o problema.

### 5. Dados acima de decoração

Gráficos e elementos visuais devem servir a uma decisão ou informação.

### 6. Consistência

O mesmo conceito deve ter a mesma aparência em todo o produto.

### 7. Open source por natureza

A interface deve transmitir transparência, controle e autonomia.

---

# 29. Design Tokens — referência rápida

```css
:root {
  --color-brand-500: #00D1B2;
  --color-brand-600: #00BFA6;
  --color-brand-700: #009B88;
  --color-brand-accent: #19E6D0;

  --color-navy-950: #0B1220;
  --color-navy-900: #0F172A;
  --color-navy-800: #172033;

  --color-white: #FFFFFF;

  --color-gray-50: #F8FAFC;
  --color-gray-100: #F1F5F9;
  --color-gray-200: #E2E8F0;
  --color-gray-300: #CBD5E1;
  --color-gray-400: #94A3B8;
  --color-gray-500: #64748B;
  --color-gray-600: #475569;
  --color-gray-700: #334155;
  --color-gray-800: #1E293B;
  --color-gray-900: #0F172A;

  --color-success: #22C55E;
  --color-warning: #F59E0B;
  --color-error: #EF4444;
  --color-info: #3B82F6;

  --radius-sm: 6px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --radius-xl: 16px;
  --radius-2xl: 20px;
  --radius-full: 9999px;
}
```

---

# 30. Resumo da identidade

**Marca:** Safora

**Categoria:** Open-source backup management

**Personalidade:**
- técnica
- confiável
- simples
- moderna
- transparente
- open source

**Símbolo:** S geométrico em forma de ribbon

**Cor principal:** `#00D1B2`

**Accent:** `#19E6D0`

**Dark background:** `#0B1220`

**Tipografia:** Outfit

**Monospace:** JetBrains Mono

**Ícones:** Lucide

**Estética:** Dark technical + clean SaaS

**Princípio central:**

> **Protect the data. Automate the rest.**
