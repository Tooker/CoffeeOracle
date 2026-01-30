"""Welcome to Reflex! This file outlines the steps to create a basic app."""

import reflex as rx

from rxconfig import config
import base64
import dotenv

dotenv.load_dotenv()

from openai import OpenAI


client = OpenAI()



class State(rx.State):
    """The app state."""
    result: str
    show: bool = False
    creativity: int = 5
    name: str
    uploaded_files: list[str] = []

    @rx.event
    async def handle_upload(
        self, files: list[rx.UploadFile]
    ):  
        for file in files:
            data = await file.read()
            path = rx.get_upload_dir() / file.name
            with path.open("wb") as f:
                f.write(data)
            self.uploaded_files.append(file.name)
            break
        
        img = base64.b64encode(data).decode("utf-8")
        self.show = True
        self.result = ""
        with client.responses.stream(
           model="gpt-5-nano",
            
            input=[
                {
                    "role": "developer",
                    "content": [
                        { "type": "input_text", "text": f"""Du bist ein Orakel und liest professionell die Zukunft aus milchschaum auf dem Kaffee.  Was bedeutet diese Tasse? Der nutzer Heißt {self.name} und er wünscht sich eine Esotherik stufe von {self.creativity
                                                                                        }""" },
                    ],
                },
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "input_image",
                            "image_url": f"data:image/jpeg;base64,{img}",
                        },
                    ],
                }
            ],
        ) as stream:

            for event in stream:
                # Inkrementeller Text
                if event.type == "response.output_text.delta":
                    print(event.delta, end="", flush=True)
                    self.result += event.delta
                    yield

                # Optional: Tool Calls, Reasoning, etc.
                elif event.type == "response.error":
                    print("Fehler:", event.error)

                elif event.type == "response.completed":
                    print("\n\n--- RESPONSE COMPLETED ---")
        
        
        

    @rx.event
    def sliderChange(self, value: list[float]):
        print(value)
        self.creativity = value

    @rx.event
    def updateName(self, value:str):
        self.name = value



def index() -> rx.Component:
    # Welcome Page (Index)
    return rx.container(
        rx.color_mode.button(position="top-right"),
        rx.vstack(
            rx.heading("Welcome to the famous Coffee Oracle!", size="9"),
            rx.upload(id="upload"),
        rx.button(
            "Upload",
            on_click=State.handle_upload(
                rx.upload_files("upload")
            ),
        ),
        rx.input(on_change=State.updateName,placeholder="Bitte Namen eingeben"),
        rx.heading(f"Select Creativity: {State.creativity}"),
        
        rx.slider(min=0,max=10,default_value=State.creativity,on_value_commit=State.sliderChange,step=1,width="30%"),
       
        rx.cond(
            State.show,
            rx.markdown(State.result)
        ),
    )
)


app = rx.App()
app.add_page(index)
