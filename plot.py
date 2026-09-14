import sys
import json
import math
import matplotlib.pyplot as plt

PALETTE = [
    "#1f77b4",  # Grupo 1: azul
    "#aec7e8",  # Grupo 2: azul claro
    "#ff7f0e",  # Grupo 3: laranja
    "#2ca02c",  # Grupo 4: verde
    "#98df8a",  # Grupo 5: verde claro
    "#d62728",  # Grupo 6: vermelho
    "#9467bd",  # Grupo 7: roxo
    "#8c564b",  # Grupo 8: marrom
    "#c5b0d5",  # Grupo 9: roxo claro
    "#e377c2",  # Grupo 10: rosa
    "#7f7f7f",  # Grupo 11: cinza
    "#c7c7c7",  # Grupo 12: cinza claro
    "#bcbd22",  # Grupo 13: verde oliva
    "#17becf",  # Grupo 14: ciano
    "#9edae5",  # Grupo 15: ciano claro
    "#ffbb78",  # Grupo 16+
    "#ff9896",
    "#c49c94",
    "#f7b6d2",
    "#dbdb8d",
]

def plot_clustering(json_path, output_png):
    with open(json_path, 'r', encoding='utf-8') as f:
        data = json.load(f)

    m_altura = data['grid_height']
    m_largura = data['grid_width']
    num_ants = data['num_ants']
    num_items = data['num_items']
    iterations = data['iterations']
    vision_radius = data['vision_radius']
    k1 = data['k1']
    k2 = data['k2']
    alpha = data['alpha']
    iteration_curr = data['iteration_current']
    phase_name = data['phase_name']
    items = data['items']
    ants = data['ants']

    fig, ax = plt.subplots(figsize=(9, 10))

    classes = sorted(list(set(item['classe'] for item in items)))

    for cls in classes:
        cls_items = [it for it in items if it['classe'] == cls]
        cols = [it['col'] for it in cls_items]
        rows = [it['row'] for it in cls_items]
        color = PALETTE[(cls - 1) % len(PALETTE)]
        ax.scatter(cols, rows, color=color, label=f"Grupo {cls}", s=35, alpha=0.9)

    if ants:
        ant_cols = [a['col'] for a in ants]
        ant_rows = [a['row'] for a in ants]
        ax.scatter(ant_cols, ant_rows, color='black', marker='^', label="Formiga", s=70, zorder=5)

    ax.set_xlim(-0.5, m_largura - 0.5)
    ax.set_ylim(m_altura - 0.5, -0.5)  # Eixo Y invertido: 0 no topo
    ax.set_xlabel("Coluna", fontsize=11)
    ax.set_ylabel("Linha", fontsize=11)

    formatted_iter = f"{iteration_curr:,}"
    ax.set_title(f"Iteração {formatted_iter} ({phase_name})", fontsize=14, fontweight='bold', pad=12)

    # Legenda organizada em colunas como na imagem de referência
    handles, labels = ax.get_legend_handles_labels()
    total_entries = len(handles)
    ncol = min(total_entries, 6)
    nrows = math.ceil(total_entries / ncol)

    # Reordena para preenchimento por coluna (column-major)
    reordered_handles = []
    reordered_labels = []
    for r in range(nrows):
        for c in range(ncol):
            idx = c * nrows + r
            if idx < total_entries:
                reordered_handles.append(handles[idx])
                reordered_labels.append(labels[idx])

    ax.legend(reordered_handles, reordered_labels, loc='upper center', bbox_to_anchor=(0.5, -0.07),
              ncol=ncol, frameon=True, fontsize=9, handletextpad=0.3, columnspacing=1.0)

    # Caixa informativa
    iter_total_str = f"{iterations:,}"
    info_text = (
        f"ALTURA MATRIZ : {m_altura} | LARGURA MATRIZ: {m_largura} | AGENTES: {num_ants} | ITENS: {num_items}\n"
        f"ITERAÇÕES: {iter_total_str} | RAIO: {vision_radius} | K1: {k1} | K2: {k2} | ALPHA: {alpha}"
    )

    plt.figtext(0.5, -0.01, info_text, ha='center', va='top', fontsize=9,
                bbox=dict(boxstyle='round,pad=0.5', facecolor='#e9ecef', edgecolor='#ced4da', alpha=0.9))

    plt.tight_layout()
    plt.savefig(output_png, bbox_inches='tight', dpi=150)
    plt.close()

if __name__ == '__main__':
    if len(sys.argv) < 3:
        print("Uso: python3 plot.py <dados.json> <saida.png>")
        sys.exit(1)
    plot_clustering(sys.argv[1], sys.argv[2])
