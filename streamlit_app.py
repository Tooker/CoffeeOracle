import base64
from io import BytesIO
from pathlib import Path

import dotenv
from openai import OpenAI
import streamlit as st
from PIL import Image


dotenv.load_dotenv()

client = OpenAI()


UPLOAD_DIR = Path("uploaded_files")
UPLOAD_DIR.mkdir(exist_ok=True)


def init_session_state() -> None:
    """Initialisiere die Session-State-Werte analog zur Reflex-State-Klasse."""

    if "result" not in st.session_state:
        st.session_state.result = ""
    if "show" not in st.session_state:
        st.session_state.show = False
    if "creativity" not in st.session_state:
        st.session_state.creativity = 5
    if "name" not in st.session_state:
        st.session_state.name = ""
    if "uploaded_files" not in st.session_state:
        st.session_state.uploaded_files = []


def handle_upload(uploaded_file) -> None:
    """Verarbeite den Upload und streame die OpenAI-Antwort in die Oberfläche."""

    init_session_state()

    if uploaded_file is None:
        st.warning("Bitte lade ein Bild deines Kaffeeschaums hoch, bevor du das Orakel befragst.")
        return

    try:
        data = uploaded_file.read()
        if not data:
            st.warning("Die hochgeladene Datei scheint leer zu sein.")
            return

        # Bild verkleinern, um Bandbreite und Kontextfenster zu schonen
        image = Image.open(BytesIO(data))
        image.thumbnail((1024, 1024))  # max. 1024x1024 Pixel
        buffer = BytesIO()
        save_format = getattr(image, "format", None) or "JPEG"
        image.save(buffer, format=save_format)
        buffer.seek(0)
        resized_data = buffer.read()

        file_path = UPLOAD_DIR / uploaded_file.name
        file_path.write_bytes(resized_data)
        st.session_state.uploaded_files.append(uploaded_file.name)

        img_b64 = base64.b64encode(resized_data).decode("utf-8")

        st.session_state.show = True
        st.session_state.result = ""

        output_placeholder = st.empty()

        mime_type = getattr(uploaded_file, "type", None) or "image/jpeg"
        image_url = f"data:{mime_type};base64,{img_b64}"

        dev_prompt = (
            "Du bist ein Orakel und liest professionell die Zukunft aus Milchschaum auf dem Kaffee. "
            "Was bedeutet diese Tasse? "
            f"Der Nutzer heißt {st.session_state.name} "
            f"und er wünscht sich eine Esotherik stufe von {st.session_state.creativity}"
        )

        with client.responses.stream(
            model="gpt-5-nano",
            input=[
                {
                    "role": "developer",
                    "content": [
                        {
                            "type": "input_text",
                            "text": dev_prompt,
                        },
                    ],
                },
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "input_image",
                            "image_url": image_url,
                        },
                    ],
                },
            ],
        ) as stream:

            for event in stream:
                # Inkrementeller Text
                if event.type == "response.output_text.delta":
                    # In der Reflex-App wird direkt event.delta verwendet.
                    delta = getattr(event, "delta", "")
                    st.session_state.result += delta
                    output_placeholder.markdown(st.session_state.result)

                # Optional: Fehlerbehandlung
                elif event.type == "response.error":
                    st.error(f"Fehler beim Aufruf des Orakels: {event.error}")

                elif event.type == "response.completed":
                    # Abschluss der Antwort
                    break

    except Exception as exc:  # pragma: no cover - robust gegen Laufzeitfehler
        st.session_state.show = False
        st.error(f"Beim Lesen deines Kaffeegeschicks ist ein Fehler aufgetreten: {exc}")


def main() -> None:
    """Baue die Streamlit-Oberfläche und binde die Upload-/OpenAI-Logik ein."""

    st.set_page_config(
        page_title="Coffee Oracle",
        page_icon="☕",
        layout="centered",
    )

    init_session_state()

   
      

    st.title("Welcome to the famous Coffee Oracle!")
    st.write(
        "Lade ein Foto deines Kaffeeschaums hoch und lasse das Orakel deine Zukunft deuten. "
        "Gib dazu deinen Namen und die gewünschte Esoterik-Stufe an."
    )
    st.text_input("Name", key="name")
    st.slider("Select Creativity", 0, 10, key="creativity")

    uploaded_file = st.file_uploader(
        "Upload your coffee foam image", type=["jpg", "jpeg", "png"]
    )

    if uploaded_file is not None:
        st.image(uploaded_file, caption="Deine Kaffeeschaum-Tasse")

    if st.button("Read my Coffee Fortune"):
        handle_upload(uploaded_file)

    #if st.session_state.show:
    #    st.markdown(st.session_state.result)


if __name__ == "__main__":  # pragma: no cover - ermöglicht lokalen Start mit `python streamlit_app.py`
    main()

