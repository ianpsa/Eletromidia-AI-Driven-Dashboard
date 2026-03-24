import "./HexbinFilters.css";

export function HexbinFilters() {
  return (
    <aside className="hexbin-filters">
      <h3 className="hexbin-filters__title">Filtros</h3>

      <label className="hexbin-filters__field">
        <span>Estado</span>
        <select>
          <option>SP</option>
          <option>RJ</option>
          <option>MG</option>
          <option>PR</option>
        </select>
      </label>

      <label className="hexbin-filters__field">
        <span>Cidade</span>
        <select>
          <option>São Paulo</option>
          <option>Campinas</option>
          <option>Santos</option>
          <option>Guarulhos</option>
          <option>Osasco</option>
        </select>
      </label>

      <label className="hexbin-filters__field">
        <span>Endereço</span>
        <select>
          <option>Rua M.M.D.C</option>
          <option>Avenida Paulista</option>
          <option>Rua da Consolação</option>
          <option>Rua Vergueiro</option>
        </select>
      </label>

      <label className="hexbin-filters__field">
        <span>Distância máxima</span>
        <input type="range" min={2} max={15} defaultValue={10} />
        <strong>10 km</strong>
      </label>

      <label className="hexbin-filters__field">
        <span>Gênero</span>
        <select>
          <option>Todos</option>
          <option>Feminino</option>
          <option>Masculino</option>
          <option>Outro</option>
        </select>
      </label>

      <label className="hexbin-filters__field">
        <span>Faixa etária</span>
        <select>
          <option>Todos</option>
          <option>18-19</option>
          <option>20-29</option>
          <option>30-39</option>
          <option>40-49</option>
          <option>50+</option>
        </select>
      </label>

      <label className="hexbin-filters__field">
        <span>Classe social</span>
        <select>
          <option>Todos</option>
          <option>Classe A/B</option>
          <option>Classe C</option>
          <option>Classe D/E</option>
        </select>
      </label>

      <div className="hexbin-filters__actions">
        <button className="hexbin-filters__primary">Aplicar filtros</button>
        <button className="hexbin-filters__secondary">Exportar gráficos</button>
      </div>
    </aside>
  );
}