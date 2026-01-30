# Coffee Oracle

Dieses Projekt enthält sowohl die ursprüngliche Reflex-App als auch eine Streamlit-Portierung des Coffee Oracles.

## Streamlit-App starten

1. Abhängigkeiten installieren (z.B. mit `uv` oder `pip`):

```bash
pip install -e .
```

2. Sicherstellen, dass in der Datei `.env` der `OPENAI_API_KEY` gesetzt ist.

3. Streamlit-App starten:

```bash
streamlit run streamlit_app.py
```

Die App lädt ein Bild deines Kaffeeschaums hoch und nutzt das OpenAI Responses API, um ein Orakel in deutscher Sprache zu streamen.
