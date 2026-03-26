import "./HexbinFilters.css";

export function HexbinFilters() {
  return (
    <aside className="hexbin-filters">
      <h3 className="hexbin-filters__title">Filtros</h3>

      <label className="hexbin-filters__field">
        <span>Estado</span>
        <select defaultValue="">
          <option value="">Todos</option>
          <option value="SP">SP</option>
          <option value="RJ">RJ</option>
          <option value="MG">MG</option>
          <option value="PR">PR</option>
        </select>
      </label>

      <label className="hexbin-filters__field">
        <span>Cidade</span>
        <select defaultValue="">
          <option value="">Todos</option>
          <option value="São Paulo">São Paulo</option>
          <option value="Campinas">Campinas</option>
          <option value="Santos">Santos</option>
          <option value="Guarulhos">Guarulhos</option>
          <option value="Osasco">Osasco</option>
        </select>
      </label>

      <label className="hexbin-filters__field">
        <span>Endereço</span>
        <select defaultValue="">
          <option value="">Todos</option>
          <option value="Rua M.M.D.C">Rua M.M.D.C</option>
          <option value="Avenida Paulista">Avenida Paulista</option>
          <option value="Rua da Consolação">Rua da Consolação</option>
          <option value="Rua Vergueiro">Rua Vergueiro</option>
        </select>
      </label>

      <label className="hexbin-filters__field">
        <span>Distância máxima</span>
        <input type="range" min={2} max={15} defaultValue={10} />
        <strong>10 km</strong>
      </label>

      <label className="hexbin-filters__field">
        <span>Horário</span>
        <select defaultValue="">
          <option value="">Todos</option>
          <option value="0-6">00:00 - 06:00</option>
          <option value="6-12">06:00 - 12:00</option>
          <option value="12-18">12:00 - 18:00</option>
          <option value="18-24">18:00 - 24:00</option>
        </select>
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
        <button type="button" className="hexbin-filters__primary">
          Aplicar filtros
        </button>
        <button type="button" className="hexbin-filters__secondary">
          Exportar gráficos
        </button>
      </div>
    </aside>
  );
}
