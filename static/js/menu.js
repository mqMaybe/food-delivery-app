document.addEventListener('DOMContentLoaded', async () => {
    const session = await checkAuth('customer');
    if (!session) return;

    const urlParams = new URLSearchParams(window.location.search);
    const restaurantId = urlParams.get('restaurant_id');

    if (!restaurantId) {
        displayError('menuList', 'ID ресторана не указан.');
        return;
    }

    try {
        const response = await fetch(`/api/menu?restaurant_id=${restaurantId}`);
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Не удалось загрузить меню');
        }

        const menuItems = await response.json();
        const container = document.getElementById('menuList');
        if (!container) {
            console.error('Container with id "menuList" not found');
            return;
        }

        if (menuItems.length === 0) {
            container.innerHTML = '<p>Меню пусто.</p>';
            return;
        }

        menuItems.forEach(item => {
            const div = document.createElement('div');
            div.className = 'menu-item';
            div.innerHTML = `
                <h3>${item.name}</h3>
                <p>${item.description}</p>
                <p>Цена: ${item.price} ₽</p>
                <button onclick="addToCart(${item.id}, ${restaurantId})">Добавить в корзину</button>
            `;
            container.appendChild(div);
        });
    } catch (error) {
        console.error('Failed to load menu:', error);
        displayError('menuList', error.message);
    }
});

async function addToCart(menuId, restaurantId) {
    try {
        const response = await fetch('/api/cart', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ menu_id: menuId, quantity: 1 }),
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Не удалось добавить в корзину');
        }

        alert('Блюдо добавлено в корзину!');
    } catch (error) {
        console.error('Failed to add to cart:', error);
        displayError('menuList', error.message);
    }
}