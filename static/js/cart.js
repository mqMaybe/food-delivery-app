// Функция для отображения ошибки
function displayError(elementId, message) {
    const element = document.getElementById(elementId);
    if (element) {
        element.innerHTML = `<p class="text-danger">${message}</p>`;
    }
}

// Загрузка корзины при загрузке страницы
document.addEventListener('DOMContentLoaded', async () => {
    console.log('cart.js загружен'); // Отладка

    const cartItems = document.getElementById('cartItems');
    const orderSummary = document.getElementById('orderSummary');
    const totalPrice = document.getElementById('totalPrice');
    const checkoutBtn = document.querySelector('.checkout-btn');
    const cartCount = document.getElementById('cartCount');

    if (!cartItems || !orderSummary || !totalPrice || !checkoutBtn || !cartCount) {
        console.error('Необходимые элементы не найдены');
        return;
    }

    const loadCart = async () => {
        try {
            console.log('Загрузка корзины...'); // Отладка
            const response = await fetch('/api/cart');
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Не удалось загрузить корзину');
            }

            const items = await response.json();
            console.log('Товары в корзине:', items); // Отладка

            cartItems.innerHTML = ''; // Очищаем список
            orderSummary.innerHTML = ''; // Очищаем итоги
            let total = 0;
            let itemCount = 0;

            if (items.length === 0) {
                cartItems.innerHTML = '<p>Корзина пуста.</p>';
                totalPrice.textContent = '$0';
                checkoutBtn.textContent = 'Оформить заказ $0';
                cartCount.textContent = '0';
                return;
            }

            items.forEach(item => {
                // Добавляем товар в список корзины
                const cartItem = document.createElement('div');
                cartItem.className = 'cart-item';
                cartItem.innerHTML = `
                    <img src="/static/images/food-placeholder.jpg" alt="${item.menu_name}">
                    <div class="cart-item-details">
                        <h5>${item.menu_name}</h5>
                        <p>${item.menu_description}</p>
                        <div class="quantity-control">
                            <button onclick="updateQuantity(${item.id}, ${item.quantity - 1})">-</button>
                            <input type="text" value="${item.quantity}" readonly>
                            <button onclick="updateQuantity(${item.id}, ${item.quantity + 1})">+</button>
                        </div>
                    </div>
                    <div>
                        <p><strong>$${item.menu_price}</strong></p>
                        <button class="remove-btn" onclick="removeItem(${item.id})">🗑️</button>
                    </div>
                `;
                cartItems.appendChild(cartItem);

                // Добавляем товар в итоги заказа
                const summaryItem = document.createElement('div');
                summaryItem.className = 'order-summary-item';
                summaryItem.innerHTML = `
                    <span>${item.quantity}x ${item.menu_name}</span>
                    <span>$${item.menu_price * item.quantity}</span>
                `;
                orderSummary.appendChild(summaryItem);

                total += item.menu_price * item.quantity;
                itemCount += item.quantity;
            });

            totalPrice.textContent = `$${total}`;
            checkoutBtn.textContent = `Оформить заказ $${total}`;
            cartCount.textContent = itemCount.toString();
        } catch (error) {
            console.error('Ошибка при загрузке корзины:', error);
            displayError('cartItems', error.message);
        }
    };

    // Обновление количества товара
    window.updateQuantity = async (itemId, newQuantity) => {
        if (newQuantity < 1) {
            removeItem(itemId);
            return;
        }

        try {
            const response = await fetch('/api/cart', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ cart_id: itemId, quantity: newQuantity }),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Не удалось обновить количество');
            }

            await loadCart(); // Перезагружаем корзину
        } catch (error) {
            console.error('Ошибка при обновлении количества:', error);
            displayError('cartItems', error.message);
        }
    };

    // Удаление товара из корзины
    window.removeItem = async (itemId) => {
        try {
            const response = await fetch('/api/cart', {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ cart_id: itemId }),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Не удалось удалить товар');
            }

            await loadCart(); // Перезагружаем корзину
        } catch (error) {
            console.error('Ошибка при удалении товара:', error);
            displayError('cartItems', error.message);
        }
    };

    // Применение промокода (заглушка)
    window.applyPromoCode = () => {
        const promoCode = document.getElementById('promoCode').value;
        if (promoCode) {
            alert('Промокод применён (функционал в разработке).');
        } else {
            alert('Пожалуйста, введите промокод.');
        }
    };

    // Оформление заказа
    window.checkout = async () => {
        try {
            const response = await fetch('/api/order', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ delivery_address: 'Укажите адрес доставки' }), // Здесь можно добавить поле для ввода адреса
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Не удалось оформить заказ');
            }

            alert('Заказ успешно оформлен!');
            window.location.href = '/my-orders';
        } catch (error) {
            console.error('Ошибка при оформлении заказа:', error);
            displayError('cartItems', error.message);
        }
    };

    // Загружаем корзину при загрузке страницы
    await loadCart();
});