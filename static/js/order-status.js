// static/js/order-status.js

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
        }

        if (!response.ok) {
            throw new Error(data.error || 'Не удалось выйти из системы');
        }

        window.location.href = '/login';
    } catch (error) {
        console.error('Ошибка при выходе из системы:', error);
        displayError(error.message);
    }
}


// Загрузка статуса заказа при загрузке страницы
document.addEventListener('DOMContentLoaded', async () => {
    console.log('order-status.js загружен'); // Отладка

    const statusSteps = document.querySelectorAll('.status-step');
    const deliveryTime = document.getElementById('deliveryTime');

    if (!statusSteps || !deliveryTime) {
        console.error('Необходимые элементы не найдены');
        return;
    }

    // Получаем order_id из URL
    const orderId = window.location.pathname.split('/').pop();
    if (!orderId || isNaN(orderId)) {
        console.error('Неверный ID заказа');
        alert('Неверный ID заказа');
        return;
    }

    try {
        console.log('Загрузка статуса заказа...'); // Отладка
        const response = await fetch(`/api/order/${orderId}`);
        console.log('Статус ответа:', response.status); // Отладка

        const contentType = response.headers.get('Content-Type');
        console.log('Content-Type:', contentType); // Отладка

        if (!response.ok) {
            const text = await response.text();
            console.log('Тело ответа (ошибка):', text); // Отладка
            let errorData;
            try {
                errorData = JSON.parse(text);
            } catch (e) {
                throw new Error(`Не удалось разобрать ответ сервера: ${text}`);
            }
            throw new Error(errorData.error || 'Не удалось загрузить статус заказа');
        }

        const order = await response.json();
        console.log('Данные заказа:', order); // Отладка

        // Обновляем статус
        statusSteps.forEach(step => {
            const stepStatus = step.getAttribute('data-step');
            if (stepStatus === order.status) {
                step.classList.add('active');
            }
        });

        // Получаем ресторан
        const restaurantResponse = await fetch(`/api/restaurants?cuisine_type=all&delivery_time=all&rating=all`);
        if (!restaurantResponse.ok) {
            const restaurantErrorText = await restaurantResponse.text();
            console.log('Ошибка загрузки ресторанов:', restaurantErrorText);
            throw new Error('Не удалось загрузить данные ресторана');
        }

        const data = await restaurantResponse.json();
        const restaurants = data.restaurants || data;

        console.log('Рестораны:', restaurants);
        if (!Array.isArray(restaurants)) {
            throw new Error('Неверный формат данных ресторана: ожидается массив');
        }

        const restaurant = restaurants.find(r => r.id === order.restaurant_id);
        if (restaurant && restaurant.delivery_time) {
            deliveryTime.textContent = `Время: ${restaurant.delivery_time} мин`;
        } else {
            deliveryTime.textContent = 'Время: не указано';
        }
    } catch (error) {
        console.error('Ошибка при загрузке статуса заказа:', error);
        alert(error.message);
    }
});

// Скрытие статуса и перенаправление
window.dismissStatus = () => {
    window.location.href = '/my-orders';
};
