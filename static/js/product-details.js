// static/js/product-details.js

// Функция для отображения ошибки
function displayError(message) {
    const errorDiv = document.getElementById('error-message');
    if (errorDiv) {
        errorDiv.innerHTML = `<p class="text-danger">${message}</p>`;
        errorDiv.style.display = 'block';
    }
}

// Функция для отображения успеха
function displaySuccess(message) {
    const successDiv = document.getElementById('success-message');
    if (successDiv) {
        successDiv.innerHTML = `<p class="text-success">${message}</p>`;
        successDiv.style.display = 'block';
    }
}

// Функция для смены главного изображения
function changeImage(src) {
    document.getElementById('main-image').src = src;
}

// Функция для обновления количества
function updateQuantity(change) {
    const quantityInput = document.getElementById('quantity');
    let quantity = parseInt(quantityInput.value, 10);
    quantity = Math.max(1, quantity + change); // Не даём количеству стать меньше 1
    quantityInput.value = quantity;
}

// Функция для добавления товара в корзину
async function orderNow(menuItemId, quantity) {
    try {
        const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        if (!csrfToken) throw new Error('CSRF-токен не найден');

        console.log('Тип menuItemId:', typeof menuItemId, 'Значение:', menuItemId); // Логируем тип и значение
        console.log('Тип quantity:', typeof quantity, 'Значение:', quantity);

        const restaurantId = parseInt(newOrderButton.getAttribute('data-restaurant-id'), 10);

        const payload = {
            menu_item_id: menuItemId,
            quantity: quantity,
            restaurant_id: restaurantId,
        };
        console.log('Отправляемый JSON:', JSON.stringify(payload)); // Логируем JSON

        const response = await fetch('/api/cart/add', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken,
            },
            body: JSON.stringify(payload),
        });

        const contentType = response.headers.get('Content-Type');
        let data = {};
        if (contentType && contentType.includes('application/json')) {
            data = await response.json();
        } else {
            const text = await response.text();
            console.error('Ответ сервера не является JSON:', text);
            throw new Error('Ответ сервера не в формате JSON');
        }

        console.log('Ответ сервера:', data);

        if (!response.ok) {
            throw new Error(data.error || 'Не удалось обновить корзину');
        }

        displaySuccess('Товар добавлен в корзину!');
    } catch (error) {
        console.error('Ошибка при добавлении в корзину:', error);
        displayError(error.message);
    }
}

// Загрузка рекомендованных блюд и установка обработчиков событий
document.addEventListener('DOMContentLoaded', async () => {
    // Загрузка рекомендованных блюд
    const dishesGrid = document.getElementById('recommended-dishes');
    try {
        const response = await fetch('/api/recommended-dishes');
        if (!response.ok) {
            throw new Error('Не удалось загрузить рекомендованные блюда');
        }

        const dishes = await response.json();
        dishesGrid.innerHTML = '';
        dishes.forEach(dish => {
            const dishId = dish.id;
            if (!dishId || isNaN(dishId)) {
                console.error('Неверный ID блюда:', dish);
                return;
            }

            const dishItem = document.createElement('div');
            dishItem.className = 'dish-item';
            dishItem.innerHTML = `
                <img src="${dish.image_url && dish.image_url.String ? dish.image_url.String : '/static/images/food-placeholder.jpg'}" alt="${dish.name}">
                <div class="dish-item-details">
                    <h3>${dish.name}</h3>
                    <p>${dish.price.toFixed(2)} ₽</p>
                </div>
            `;
            dishItem.addEventListener('click', () => {
                window.location.href = `/product-details/${dishId}`;
            });
            dishesGrid.appendChild(dishItem);
        });
    } catch (error) {
        console.error('Ошибка при загрузке рекомендованных блюд:', error);
        dishesGrid.innerHTML = `<p class="text-danger">${error.message}</p>`;
    }

    // Загрузка отзывов (пока статические)
    const reviewsList = document.getElementById('reviews-list');
    reviewsList.innerHTML = `
        <div class="review-item">
            <div class="stars">★★★★★</div>
            <p>Вкуснейшая паста! Очень рекомендую.</p>
            <p class="author">— Анна К.</p>
        </div>
        <div class="review-item">
            <div class="stars">★★★★☆</div>
            <p>Всё отлично, но доставка немного задержалась.</p>
            <p class="author">— Иван П.</p>
        </div>
    `;

    // Добавляем обработчик события для кнопки "Заказать сейчас"
    const orderButton = document.querySelector('.order-now-btn');
    if (orderButton) {
        orderButton.replaceWith(orderButton.cloneNode(true));
        const newOrderButton = document.querySelector('.order-now-btn');
        newOrderButton.addEventListener('click', async (event) => {
            event.preventDefault();
            const menuItemId = parseInt(newOrderButton.getAttribute('data-item-id'), 10);
            const quantityInput = document.getElementById('quantity');
            const quantity = quantityInput ? parseInt(quantityInput.value, 10) : 1;
            await orderNow(menuItemId, quantity);
            displaySuccess("Товар успешно добавлен в корзину!")
            window.location.href = '/cart';
        });
    }
});