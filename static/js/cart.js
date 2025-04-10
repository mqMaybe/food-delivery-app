// static/js/cart.js

async function logout() {
    try {
        const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        if (!csrfToken) throw new Error('CSRF-токен не найден');

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
        displayError('menu-grid', error.message);
    }
}

document.addEventListener('DOMContentLoaded', async () => {
    const cartItemsContainer = document.getElementById('cart-items');
    const orderSummaryItems = document.getElementById('order-summary-items');
    const cartTotal = document.getElementById('cart-total');
    const checkoutButton = document.getElementById('checkoutButton');

    if (!cartItemsContainer || !orderSummaryItems || !cartTotal || !checkoutButton) {
        console.error('Не удалось найти элементы на странице');
        return;
    }

    checkoutButton.addEventListener('click', checkout);

    try {
        const response = await fetch('/api/cart', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('auth_token') || ''}`,
            },
        });

        if (!response.ok) {
            throw new Error('Не удалось загрузить корзину');
        }

        const cart = await response.json();
        console.log('Корзина:', cart);

        if (cart.length === 0) {
            cartItemsContainer.innerHTML = '';
            orderSummaryItems.innerHTML = '';
            cartTotal.innerHTML = '';
            checkoutButton.style.display = 'none';
            return;
        }

        cartItemsContainer.innerHTML = '';
        orderSummaryItems.innerHTML = '';
        let total = 0;

        cart.forEach(item => {
            const itemTotal = item.price * item.quantity;
            total += itemTotal;

            const cartItem = document.createElement('div');
            cartItem.className = 'cart-item';
            cartItem.innerHTML = `
                <img src="${item.image_url && item.image_url.String ? item.image_url.String : '/static/images/food-placeholder.jpg'}" alt="${item.name}">
                <div class="cart-item-details">
                    <h5>${item.name}</h5>
                    <p>Цена: ${item.price.toFixed(2)} ₽</p>
                    <div class="quantity-control">
                        <button onclick="updateQuantity(${item.id}, -1)">−</button>
                        <input type="text" value="${item.quantity}" readonly>
                        <button onclick="updateQuantity(${item.id}, 1)">+</button>
                    </div>
                    <p>Итого: ${itemTotal.toFixed(2)} ₽</p>
                    <button class="remove-btn" onclick="removeFromCart(${item.id})">Удалить</button>
                </div>
            `;
            cartItemsContainer.appendChild(cartItem);

            const summaryItem = document.createElement('div');
            summaryItem.className = 'order-summary-item';
            summaryItem.innerHTML = `
                <span>${item.name} x${item.quantity}</span>
                <span>${itemTotal.toFixed(2)} ₽</span>
            `;
            orderSummaryItems.appendChild(summaryItem);
        });

        cartTotal.innerHTML = `
            <span>Общая сумма:</span>
            <span>${total.toFixed(2)} ₽</span>
        `;
        checkoutButton.style.display = 'block';
    } catch (error) {
        console.error('Ошибка при загрузке корзины:', error);
        cartItemsContainer.innerHTML = `<p class="text-danger">${error.message}</p>`;
    }
});

async function updateQuantity(itemId, change) {
    try {
        const authToken = localStorage.getItem('auth_token') || '';

        const response = await fetch(`/api/cart/${itemId}`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]').getAttribute('content'),
                'Authorization': `Bearer ${authToken}`,
            },
            body: JSON.stringify({ change }),
        });

        if (!response.ok) {
            const errorData = await response.json();
            if (errorData.error === 'Количество должно быть больше 0') {
                const confirmDelete = confirm('Количество станет 0. Хотите удалить товар из корзины?');
                if (confirmDelete) {
                    await removeFromCart(itemId);
                }
                return;
            }
            throw new Error(errorData.error || 'Не удалось обновить количество');
        }

        window.location.reload();
    } catch (error) {
        console.error('Ошибка при обновлении количества:', error);
        alert(error.message);
    }
}

async function removeFromCart(itemId) {
    try {
        const authToken = localStorage.getItem('auth_token') || '';

        const response = await fetch(`/api/cart/${itemId}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]').getAttribute('content'),
                'Authorization': `Bearer ${authToken}`,
            },
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Не удалось удалить товар из корзины');
        }

        window.location.reload();
    } catch (error) {
        console.error('Ошибка при удалении товара:', error);
        alert(error.message);
    }
}

async function checkout() {
    try {
        const authToken = localStorage.getItem('auth_token') || '';
        const promoCode = document.getElementById('promoCode').value || '';
        const deliveryAddress = document.getElementById('deliveryAddress').value.trim();

        // Валидация адреса доставки
        if (!deliveryAddress) {
            throw new Error('Пожалуйста, укажите адрес доставки');
        }

        const response = await fetch('/api/checkout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]').getAttribute('content'),
                'Authorization': `Bearer ${authToken}`,
            },
            body: JSON.stringify({ promoCode, deliveryAddress }),
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Не удалось оформить заказ');
        }

        alert('Заказ успешно оформлен!');
        window.location.href = '/my-orders';
    } catch (error) {
        console.error('Ошибка при оформлении заказа:', error);
        alert(error.message);
    }
}