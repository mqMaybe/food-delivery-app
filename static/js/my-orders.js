// static/js/my-orders.js

// Функция для отображения ошибки
function displayError(elementId, message) {
    const element = document.getElementById(elementId);
    if (element) {
        element.innerHTML = `<p class="text-danger">${message}</p>`;
    }
}

// Функция для выхода из системы
async function logout() {
    try {
        const response = await fetch('/api/logout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.error || 'Не удалось выйти из системы');
        }

        window.location.href = '/login';
    } catch (error) {
        console.error('Ошибка при выходе из системы:', error);
        displayError('orderList', error.message);
    }
}

// Загрузка заказов при загрузке страницы
document.addEventListener('DOMContentLoaded', async () => {
    console.log('my-orders.js загружен'); // Отладка

    const orderList = document.getElementById('orderList');
    if (!orderList) {
        console.error('Элемент списка заказов не найден');
        return;
    }

    const loadOrders = async () => {
        try {
            console.log('Загрузка заказов...'); // Отладка
            const response = await fetch('/api/orders');
            console.log('Статус ответа:', response.status); // Отладка
            if (!response.ok) {
                const errorData = await response.json();
                console.error('Ошибка в ответе:', errorData); // Отладка
                throw new Error(errorData.error || 'Не удалось загрузить заказы');
            }

            const orders = await response.json();
            console.log('Заказы:', orders); // Отладка

            orderList.innerHTML = ''; // Очищаем список
            if (orders.length === 0) {
                orderList.innerHTML = '<p>У вас пока нет заказов.</p>';
                return;
            }

            orders.forEach(order => {
                console.log('Обрабатываем заказ:', order); // Отладка
                const orderCard = document.createElement('div');
                orderCard.className = 'order-card';
                orderCard.innerHTML = `
                    <h5>Заказ #${order.id}</h5>
                    <p><strong>Адрес доставки:</strong> ${order.delivery_address}</p>
                    <p><strong>Статус:</strong> ${getStatusText(order.status)}</p>
                    <p><strong>Итоговая сумма:</strong> $${order.total_price.toFixed(2)}</p>
                    <div class="order-items">
                        <p><strong>Товары:</strong></p>
                        <ul>
                            ${order.items.map(item => `
                                <li>${item.quantity}x ${item.menu_name} - $${(item.menu_price * item.quantity).toFixed(2)}</li>
                            `).join('')}
                        </ul>
                    </div>
                    <a href="/order-status/${order.id}" class="track-btn">Отследить заказ</a>
                `;
                orderList.appendChild(orderCard);
            });
        } catch (error) {
            console.error('Ошибка при загрузке заказов:', error);
            displayError('orderList', error.message);
        }
    };

    // Функция для перевода статуса на русский
    function getStatusText(status) {
        switch (status) {
            case 'preparing':
                return 'Готовится';
            case 'en_route':
                return 'В пути';
            case 'delivered':
                return 'Доставлен';
            default:
                return 'Неизвестный статус';
        }
    }

    // Загружаем заказы при загрузке страницы
    await loadOrders();
});