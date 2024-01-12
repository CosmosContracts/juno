import base64
import json

with open("web.zip", "rb") as f:
    bytes = f.read()
    encoded = base64.b64encode(bytes)

with open("web.zip.txt", "wb") as f:
    f.write(encoded)

with open("new_web_cmd.txt", "w") as f:
    wasm_execute_json = json.dumps({"new_website": {"name": "test", "source": encoded.decode("utf-8")}})
    f.write(wasm_execute_json.replace(" ", ""))