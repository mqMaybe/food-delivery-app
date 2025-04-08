document.addEventListener('DOMContentLoaded', async () => {
    const session = await checkAuth('restaurant');
    if (!session) return;

    const loadOrdersBtn = document.getElementById('loadRestaurantOrdersBtn');
    const restaurantIdInput = document.getElementById('restaurantIdOrders');
    const ordersList = document.getElementById('restaurantOrdersList');

    if (!loadOrdersBtn || !restaurantIdInput || !ordersList) {
        console.error('Required elements not found');
        return;
    }

    loadOrdersBtn.addEventListener('click', async () => {
        const restaurantId = restaurantIdInput.value;
        if (!restaurantId) {
            displayError('restaurantOrdersList', 'Укажите ID ресторана');
            return;
        }

        try {
            const response = await fetch(`/api/restaurant/orders?restaurant_id=${restaurantId}`);
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Не удалось загрузить заказы');
            }

            const orders = await response.json();
            if (orders.length === 0) {
                ordersList.innerHTML = '<p>Заказов нет.</p>';
                return;
            }

            ordersList.innerHTML = '';
            orders.forEach(order => {
                const div = document.createElement('div');
                let itemsHtml = order.items.map(item => `
                    <li>${item.menu_name} - ${item.menu_price} ₽ x ${item.quantity}</li>
                `).join('');
                div.innerHTML = `
                    <p>Заказ #${order.id} | Статус: ${order.status}</p>
                    <p>Адрес доставки: ${order.delivery_address}</p>
                    <p>Итого: ${order.total_price} ₽</p>
                    <ul>${itemsHtml}</ul>
                    <select onchange="updateOrderStatus(${order.id}, ${restaurantId}, this.value)">
                        <option value="pending" ${order.status === 'pending' ? 'selected' : ''}>В ожидании</option>
                        <option value="preparing" ${order.status === 'preparing' ? 'selected' : ''}>Готовится</option>
                        <option value="delivered" ${order.status === 'delivered' ? 'selected' : ''}>Доставлен</option>
                    </select>
                `;
                ordersList.appendChild(div);
            });
        } catch (error) {
            console.error('Failed to load orders:', error);
            displayError('restaurantOrdersList', error.message);
        }
    });
});

async function updateOrderStatus(orderId, restaurantId, status) {
    try {
        const response = await fetch('/api/restaurant/orders', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ order_id: orderId, restaurant_id: restaurantId, status }),
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Не удалось обновить статус заказа');
        }

        alert('Статус заказа обновлён!');
    } catch (error) {
        console.error('Failed to update order status:', error);
        displayError('restaurantOrdersList', error.message);
    }
}