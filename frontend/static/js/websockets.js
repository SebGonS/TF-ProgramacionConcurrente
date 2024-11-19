let ws = new WebSocket("ws://localhost:5000/ws");

ws.onopen = () => {
    console.log("Conexión con WebSockets abierta");
};

ws.onmessage = (event) => {
    const recommendationList = document.getElementById("recommendationList");
    recommendationList.innerHTML = ""; // limpiar recomendaciones previas

    const recommendations = JSON.parse(event.data);
    if (recommendations.length > 0) {
        recommendations.forEach(product => {
            const listItem = document.createElement("li");
            listItem.classList.add("list-group-item");
            listItem.textContent = product;
            recommendationList.appendChild(listItem);
        });
    } else {
        const listItem = document.createElement("li");
        listItem.classList.add("list-group-item", "text-danger");
        listItem.textContent = "No se encontraron recomendaciones";
        recommendationList.appendChild(listItem);
    }
};

ws.onclose = () => {
    console.log("Conexión cerrada");
};

document.getElementById("sendButton").addEventListener("click", () => {
    const productSelect = document.getElementById("productSelect");
    const selectedProduct = productSelect.value;
    if (selectedProduct) {
        ws.send(selectedProduct);
    } else {
        alert("Seleccionar un producto primero");
    }
});
