# Relatório Técnico — Ant Clustering com Dados Reais e Paralelização

## 1. Visão Geral

Este projeto implementa o algoritmo de **agrupamento por formigas** (*Ant Clustering*) baseado no modelo de **Lumer & Faieta (1994)**, aplicado ao agrupamento de dados multidimensionais reais. A ideia central é simular o comportamento coletivo de formigas que, sem comunicação direta entre si, conseguem organizar "objetos" em grupos coesos emergindo apenas de regras locais simples.

Cada item do dataset ocupa uma célula em uma grade 2D toroidal e as formigas caminham pela grade decidindo probabilisticamente se pegam ou soltam itens com base na **similaridade local** dos vizinhos.

---

## 2. Representação dos Dados

### 2.1 Estrutura Dado

Cada item do dataset é representado pela struct `Dado`:

```go
type Dado struct {
    atributos []float64   // valores normalizados em [0, 1]
    classe    int         // rótulo da classe original (1..4 ou 1..15)
}
```

### 2.2 Normalização Min-Max

Antes de posicionar os dados na grade, todos os atributos são normalizados por **Min-Max** individualmente por dimensão:

$$x_{norm} = \frac{x - x_{min}}{x_{max} - x_{min}} \in [0, 1]$$

Isso garante que nenhuma dimensão domine o cálculo de distância por diferença de escala.

### 2.3 Preenchimento da Grade

Os $N$ dados normalizados são distribuídos **aleatoriamente** em $N$ células distintas de uma grade $M \times M$ (usando permutação aleatória), com as demais células vazias (`nil`).

---

## 3. Heurística de Agrupamento (Lumer & Faieta)

### 3.1 Distância Euclidiana

A similaridade entre dois dados $i$ e $j$ é medida pela distância euclidiana entre seus vetores de atributos:

$$d(x_i, x_j) = \sqrt{\sum_{k=1}^{D} (x_{i,k} - x_{j,k})^2}$$

### 3.2 Função de Densidade de Vizinhança $f(x_i)$

Para cada formiga na posição $(x, y)$, a função de densidade mede o quanto os vizinhos no raio $r$ se assemelham ao item $x_i$ sendo avaliado:

$$f(x_i) = \max\left(0,\ \frac{1}{s^2} \sum_{x_j \in \text{vizinhança}} \left(1 - \frac{d(x_i, x_j)}{\alpha}\right)\right)$$

Onde:
- $s = 2r + 1$ → tamanho da janela de visão (ex.: $r=4 \Rightarrow s=9 \Rightarrow s^2=81$ células)
- $\alpha = 0{,}5$ → limiar de distância; se $d(x_i, x_j) > \alpha$, o vizinho contribui negativamente
- A soma é clampeada em $[0, 1]$

**Interpretação:** $f(x_i) \approx 1$ quando o item está rodeado por vizinhos muito similares (cluster coeso); $f(x_i) \approx 0$ quando está isolado ou rodeado por itens diferentes.

### 3.3 Probabilidade de Pegar um Item

Quando uma formiga desocupada encontra um item, ela o pega com probabilidade:

$$P_p(x_i) = \left(\frac{k_1}{k_1 + f(x_i)}\right)^2, \quad k_1 = 0{,}1$$

| Situação | $f(x_i)$ | $P_p$ |
|---|---|---|
| Item isolado (sem vizinhos similares) | $\approx 0$ | $\approx 1{,}0$ (pega quase sempre) |
| Item em cluster coeso | $\approx 1$ | $\approx 0{,}01$ (quase nunca pega) |

A formiga tende a **desmantelar itens isolados** e **preservar clusters** já formados.

### 3.4 Probabilidade de Largar um Item

Quando uma formiga carregando um item encontra uma célula vazia, ela solta com probabilidade:

$$P_d(x_i) = \left(\frac{f(x_i)}{k_2 + f(x_i)}\right)^2, \quad k_2 = 0{,}15$$

| Situação | $f(x_i)$ | $P_d$ |
|---|---|---|
| Vizinhança sem similares | $\approx 0$ | $\approx 0$ (não larga) |
| Vizinhança rica em similares | $\approx 1$ | $\approx 0{,}79$ (larga com alta prob.) |

A formiga tende a **depositar itens onde já há vizinhos similares**, reforçando clusters.

---

## 4. Impacto do Raio de Visão $r$

O raio de visão define o tamanho da janela onde a formiga avalia a vizinhança para calcular $f(x_i)$.

| Raio $r$ | Janela $(2r+1)^2$ | Células vizinhas | % da grade 50×50 | Efeito |
|---|---|---|---|---|
| 1 | 3×3 | 8 | 0,3% | Visão muito local; $f$ quase sempre ≈ 0 |
| 2 | 5×5 | 24 | 1% | Sinal fraco; convergência lenta |
| **3** | 7×7 | 48 | 2% | Bom balanço local |
| **4** | **9×9** | **80** | **3,2%** | **Recomendado (literatura)** |
| 5 | 11×11 | 120 | 4,8% | Bom, começa a perder localidade |
| ≥6 | ≥13×13 | ≥168 | ≥6,7% | Sinal muito global; clusters confundidos |

### Por que raio 1 falha?

Com $r=1$ e ocupação de ~16% (400 itens em 2500 células), em média apenas **1,3 vizinhos** são encontrados na janela 3×3. O $f(x_i)$ fica sempre próximo de 0 para qualquer item, fazendo com que:
- $P_p \approx 1$: a formiga pega quase tudo que encontra
- $P_d \approx 0$: a formiga raramente solta

O resultado é **caos**: as formigas ficam carregando itens indefinidamente sem conseguir formar grupos.

### Por que raio 4 funciona?

Com $r=4$, a janela 9×9 cobre 80 células. Com 16% de ocupação, em média **~13 vizinhos** são encontrados. O sinal de $f(x_i)$ tem variação suficiente para distinguir itens isolados de itens em clusters, guiando o agrupamento emergente de forma eficaz.

---

## 5. Parâmetros Utilizados

| Parâmetro | Valor | Descrição |
|---|---|---|
| $\alpha$ | 0,5 | Limiar de similaridade na função $f$ |
| $k_1$ | 0,1 | Controla a sensibilidade de pegar ($P_p$ alta quando $f$ baixo) |
| $k_2$ | 0,15 | Controla a sensibilidade de largar ($P_d$ alta quando $f$ alto) |
| Grade | 50×50 | Ambiente toroidal |
| Agentes | 15 (dataset 4) / 20 (dataset 15) | Formigas ativas na simulação |
| Raio | 4 | Janela de visão 9×9 |
| Iterações | 2.500.000 | Passos totais da simulação |

---

## 6. Visualizações da Simulação

As imagens abaixo mostram a evolução do agrupamento ao longo da simulação (dataset `4.txt`, 400 itens, 4 classes):

````carousel
![Distribuição Inicial — iteração 0](/home/andre/.gemini/antigravity-ide/brain/429629db-8400-4a0f-953d-201c04847a0d/inicio.png)
<!-- slide -->
![Distribuição Intermediária — iteração 1.250.000](/home/andre/.gemini/antigravity-ide/brain/429629db-8400-4a0f-953d-201c04847a0d/meio.png)
<!-- slide -->
![Distribuição Final — iteração 2.500.000](/home/andre/.gemini/antigravity-ide/brain/429629db-8400-4a0f-953d-201c04847a0d/final.png)
````

---

## 7. Paralelização com Metodologia PCAM

A versão paralela do algoritmo ([parallel-main.go](file:///home/andre/udesc/ia/ant-clustering/parallel-main.go)) foi projetada utilizando a metodologia **PCAM** (Partitioning, Communication, Agglomeration, Mapping) de Ian Foster.

### 7.1 Partitioning — Decomposição por Agentes

Cada formiga é tratada como uma **tarefa computacional independente**. As operações por tarefa em cada passo são:
1. Leitura da vizinhança para cálculo de $f(x_i)$
2. Sorteio probabilístico de pegar/largar
3. Atualização da célula (escrever `nil` ou `*Dado`)
4. Movimentação com verificação de colisão

### 7.2 Communication — Pontos de Contenção

| Recurso compartilhado | Acesso | Solução |
|---|---|---|
| `matriz[nx][ny]` (leitura de vizinhos em `calc_f`) | Múltiplos leitores simultâneos | `sync.RWMutex` por célula — `RLock` para leitura |
| `matriz[x][y]` (escrever pegar/largar) | Escrita exclusiva na célula atual | `cellMutex[x][y].Lock()` com verificação dupla (*double-check*) |
| `ocupada[nx][ny]` (movimentação) | Leitura+Escrita exclusiva | `sync.Mutex` global para movimentação |

### 7.3 Agglomeration — Loteamento de Iterações

Para evitar criar **milhões de goroutines** (uma por iteração por formiga), as iterações são agrupadas em **lotes de 1000 passos** (`batchSize = 1000`):

```
para cada bloco de 1000 iterações:
    cria numAgentes goroutines, cada uma executa 1000 passos
    wg.Wait()  ← sincroniza antes do próximo lote
próximo lote...
```

Resultado: ao invés de $2{,}5 \times 10^6 \times N_{\text{formigas}}$ criações de goroutine, apenas $\frac{2{,}5 \times 10^6}{1000} \times N_{\text{formigas}} = 2500 \times N_{\text{formigas}}$ — **1000 vezes menos overhead**.

### 7.4 Mapping — Mapeamento para Núcleos da CPU

As goroutines são escalonadas pelo runtime do Go automaticamente nos núcleos físicos disponíveis. Cada goroutine mantém seu próprio gerador pseudo-aleatório (`rand.New(rand.NewSource(...))`) para eliminar contenção no PRNG global.

### 7.5 Análise de Desempenho: Raio × Eficiência da Paralelização

| Configuração | Sequencial | Paralelo | Speedup |
|---|---|---|---|
| 50 agentes, raio=1, 2,5M iter | ~11s | ~25s | 0,44× (mais lento) |
| 50 agentes, raio=4, 2,5M iter | ~47s | ~26s | **1,8×** |
| 15 agentes, raio=4, 2,5M iter | ~11s | ~8s | **1,4×** |

**Conclusão:** o benefício da paralelização é proporcional ao custo computacional do `calc_f`. Com raio pequeno ($r=1$, 8 vizinhos), o trabalho por passo é ínfimo e o overhead de sincronização de mutexes domina. Com raio maior ($r=4$, 80 vizinhos com cálculo euclidiano), o paralelismo compensa. O ponto de equilíbrio está em **raio ≥ 3-4 com pelo menos 10 agentes**.

### 7.6 Otimizações Implementadas

1. **Eliminação de alocações no heap:** substituição de `rand.Perm(4)` (que aloca um slice no heap a cada passo) por embaralhamento in-place em array de tamanho fixo `[4]int` na pilha — eliminando ~125 milhões de alocações em 2,5M iterações com 50 formigas.

2. **Lock granular no movimento:** ao invés de `defer mutex.Unlock()` (que segura a trava até o retorno da função), o lock é adquirido e liberado imediatamente após cada verificação de célula vizinha.

3. **Geração de gráficos assíncrona:** ao salvar o estado da grade (Início, Meio, Final), a renderização Python é disparada em uma goroutine separada controlada por `sync.WaitGroup`, permitindo que a simulação continue sem bloquear.

---

## 8. Referências

- LUMER, E. D.; FAIETA, B. *Diversity and Adaptation in Populations of Clustering Ants*. Proceedings of the Third International Conference on Simulation of Adaptive Behavior, 1994.
- DENEUBOURG, J. L. et al. *The Dynamics of Collective Sorting: Robot-like Ants and Ant-like Robots*. Proceedings of the First International Conference on Simulation of Adaptive Behavior, 1991.
- FOSTER, I. *Designing and Building Parallel Programs*. Addison-Wesley, 1995. (Metodologia PCAM)
