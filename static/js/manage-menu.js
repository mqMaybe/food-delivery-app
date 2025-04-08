document.addEventListener('DOMContentLoaded', async () => {
    const session = await checkAuth('restaurant');
    if (!session) return;

    const loadMenuBtn = document.getElementById('loadMenuBtn');
    const restaurantIdInput = document.getElementById('manageRestaurantId');
    const manageMenuList = document.getElementById('manageMenuList');

    if (!loadMenuBtn || !restaurantIdInput || !manageMenuList) {
        console.error('Required elements not found');
        return;
    }

    loadMenuBtn.addEventListener('click', async () => {
        const restaurantId = restaurantIdInput.value;
        if (!restaurantId) {
            displayError('manageMenuList', 'Укажите ID ресторана');
            return;
        }

        try {
            const response = await fetch(`/api/menu?restaurant_id=${restaurantId}`);
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Не удалось загрузить меню');
            }

            const menuItems = await response.json();
            if (menuItems.length === 0) {
                manageMenuList.innerHTML = '<p>Меню пусто. Добавьте блюда.</p>';
                return;
            }

            manageMenuList.innerHTML = '';
            menuItems.forEach(item => {
                const div = document.createElement('div');
                div.innerHTML = `
                    <p>ID: ${item.id} | ${item.name} - ${item.price} ₽: ${item.description}</p>
                    <button onclick="editMenuItem(${item.id}, ${restaurantId})">Редактировать</button>
                    <button onclick="deleteMenuItem(${item.id}, ${restaurantId})">Удалить</button>
                `;
                manageMenuList.appendChild(div);
            });
        } catch (error) {
            console.error('Failed to load menu:', error);
            displayError('manageMenuList', error.message);
        }
    });
});

async function editMenuItem(id, restaurantId) {
    const name = prompt('Введите новое название:', '');
    const price = prompt('Введите новую цену:', '');
    const description = prompt('Введите новое описание:', '');

    if (!name || !price || !description) {
        displayError('manageMenuList', 'Все поля обязательны');
        return;
    }

    try {
        const response = await fetch('/api/menu', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                menu_id: id,
                restaurant_id: restaurantId,
                name,
                price: parseFloat(price),
                description,
            }),
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Не удалось обновить блюдо');
        }

        window.location.reload();
    } catch (error) {
        console.error('Failed to edit menu item:', error);
        displayError('manageMenuList', error.message);
    }
}

async function deleteMenuItem(id, restaurantId) {
    if (!confirm('Вы уверены, что хотите удалить это блюдо?')) {
        return;
    }

    try {
        const response = await fetch('/api/menu', {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ menu_id: id, restaurant_id: restaurantId }),
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Не удалось удалить блюдо');
        }

        window.location.reload();
    } catch (error) {
        console.error('Failed to delete menu item:', error);
        displayError('manageMenuList', error.message);
    }
}