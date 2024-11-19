# from flask import Flask, render_template, request, jsonify
# from flask_sock import Sock

# app = Flask(__name__)
# sock = Sock(app)

# tasks = []

# @app.route('/')
# def index():
#     return render_template('index.html')

# @app.route('/tasks', methods=['POST'])
# def create_task():
#     data = request.get_json()
#     tasks.append(data)
#     return jsonify({"status": "Task created"}), 201

# @sock.route('/ws')
# def websocket(ws):
#     while True:
#         data = ws.receive()
#         ws.send(f"Received: {data}")

# if __name__ == '__main__':
#     app.run(debug=True)
from flask import Flask, render_template, request, jsonify
from flask_sock import Sock

app = Flask(__name__)
sock = Sock(app)

# Example list of products (can be replaced with actual data or fetched from a database)
products = ["Product A", "Product B", "Product C", "Product D", "Product E"]

# Dummy recommendations (for testing purposes)
recommendations = {
    "Product A": ["Product B", "Product C"],
    "Product B": ["Product A", "Product D"],
    "Product C": ["Product A", "Product E"],
    "Product D": ["Product B", "Product E"],
    "Product E": ["Product C", "Product D"]
}

@app.route('/')
def index():
    return render_template('index.html', products=products)

@sock.route('/ws')
def websocket(ws):
    while True:
        product_name = ws.receive()
        if product_name in recommendations:
            # Send back the list of recommended products as a JSON string
            ws.send(jsonify(recommendations[product_name]).get_data(as_text=True))
        else:
            ws.send(jsonify([]).get_data(as_text=True))

if __name__ == '__main__':
    app.run(debug=True)
