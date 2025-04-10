// static/js/restaurant-orders.js

// Функция для отображения ошибки
function displayError(elementId, message) {
    const element = document.getElementById(elementId);
    if (element) {
        element.innerHTML = `<p class="text-danger">${message}</p>`;
        element.style.display = 'block';
    } else {
        console.error(`Element with ID ${elementId} not found`);
    }
}

// Функция для выхода из системы
async function logout() {
    try {
        const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        if (!csrfToken) {
            throw new Error('CSRF-токен не найден');
        }

        const response = await fetch('/api/logout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken,
            },
        });

        const contentType = response.headers.get('Content-Type');
        let data = {};
        if (contentType && contentType.includes('application/json')) {
            data = await response.json();
        } else {
            const text = await response.text();
            console.error('Ответ сервера не является JSON:', text);
            throw new Error('Ответ сервера не является JSON');
        }

        if (!response.ok) {
            throw new Error(data.error || 'Не удалось выйти из системы');
        }

        window.location.href = '/login';
    } catch (error) {
        console.error('Ошибка при выходе из системы:', error);
        displayError('restaurantOrdersList', error.message);
    }
}

document.addEventListener('DOMContentLoaded', async () => {

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
                div.className = 'order-item';
                // Проверяем, есть ли order.items, и если нет — показываем пустой список
                const itemsHtml = (order.items && Array.isArray(order.items))
                    ? order.items.map(item => `
                        <li>${item.menu_name} - ${item.menu_price} ₽ x ${item.quantity}</li>
                    `).join('')
                    : '<li>Элементы заказа отсутствуют.</li>';
                div.innerHTML = `
                    <p>Заказ #${order.id} | Статус: ${order.status}</p>
                    <p>Адрес доставки: ${order.delivery_address}</p>
                    <p>Итого: ${order.total_price} ₽</p>
                    <select onchange="updateOrderStatus(${order.id}, ${restaurantId}, this.value)">
                        <option value="pending" ${order.status === 'pending' ? 'selected' : ''}>Не оплачен</option>
                        <option value="preparing" ${order.status === 'preparing' ? 'selected' : ''}>Готовится</option>
                        <option value="en_route" ${order.status === 'en_route' ? 'selected' : ''}>В пути</option>
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
        const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        if (!csrfToken) {
            throw new Error('CSRF-токен не найден');
        }

        const response = await fetch('/api/restaurant/orders', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken,
            },
            body: JSON.stringify({ order_id: orderId, restaurant_id: restaurantId, status }),
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Не удалось обновить статус заказа');
        }

        showToast('Статус заказа обновлён!');
    } catch (error) {
        console.error('Failed to update order status:', error);
        displayError('restaurantOrdersList', error.message);
    }
}

function showToast(message, isError = false) {
    const toast = document.createElement('div');
    toast.className = `toast ${isError ? 'error' : ''}`;
    toast.textContent = message;
    document.getElementById('toastContainer').appendChild(toast);

    setTimeout(() => {
        toast.classList.add('show');
    }, 100);

    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}