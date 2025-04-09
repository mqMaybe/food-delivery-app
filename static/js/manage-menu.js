// static/js/manage-menu.js

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
        displayError('manageMenuList', error.message);
    }
}

document.addEventListener('DOMContentLoaded', async () => {
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
            const response = await fetch(`/api/menu-restaurants?restaurant_id=${restaurantId}`);
            let responseBody;

            // Проверяем заголовок Content-Type
            const contentType = response.headers.get('Content-Type');
            if (!contentType || !contentType.includes('application/json')) {
                responseBody = await response.text();
                console.error('Ответ сервера не является JSON:', responseBody);
                throw new Error('Ответ сервера не является JSON');
            }

            // Парсим JSON только один раз
            responseBody = await response.json();

            if (!response.ok) {
                throw new Error(responseBody.error || `Ошибка сервера: ${response.status} ${response.statusText}`);
            }

            const menuItems = responseBody;
            if (menuItems.length === 0) {
                manageMenuList.innerHTML = '<p>Меню пусто. Добавьте блюда.</p>';
                return;
            }

            manageMenuList.innerHTML = '';
            menuItems.forEach(item => {
                const div = document.createElement('div');
                div.className = 'menu-item';
                div.innerHTML = `
                    <p>ID: ${item.id} | ${item.name} - ${item.price} ₽</p>
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
        const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        if (!csrfToken) {
            throw new Error('CSRF-токен не найден');
        }

        const response = await fetch('/api/menu', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken,
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
        const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        if (!csrfToken) {
            throw new Error('CSRF-токен не найден');
        }

        const payload = {
            menu_id: id,
            restaurant_id: restaurantId,
        };
        console.log('Отправляемый JSON:', JSON.stringify(payload));

        const response = await fetch('/api/menu', {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken,
            },
            body: JSON.stringify(payload),
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