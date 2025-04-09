// Функция для получения параметра из URL
function getQueryParam(name) {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get(name);
}

// Функция для отображения ошибки
function displayError(elementId, message) {
    const element = document.getElementById(elementId);
    if (element) {
        element.innerHTML = `<p class="text-danger">${message}</p>`;
        element.style.display = 'block';
    }
}

// Функция для выхода из системы
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

// Функция для обновления количества
function updateQuantity(itemId, change) {
    const quantityControl = document.querySelector(`.quantity-control[data-item-id="${itemId}"]`);
    if (quantityControl) {
        const input = quantityControl.querySelector('input');
        if (input) {
            let quantity = parseInt(input.value, 10);
            quantity = Math.max(1, quantity + change);
            input.value = quantity;
        }
    }
}

// Загрузка данных о ресторане и меню
document.addEventListener('DOMContentLoaded', async () => {
    console.log('menu.js загружен');

    const restaurantId = getQueryParam('restaurant_id');
    if (!restaurantId) {
        displayError('menu-grid', 'ID ресторана не указан в URL');
        return;
    }

    try {
        const response = await fetch(`/api/menu?restaurant_id=${restaurantId}`);
        const contentType = response.headers.get('Content-Type');

        if (!response.ok) {
            if (contentType && contentType.includes('application/json')) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Ошибка сервера');
            } else {
                const text = await response.text();
                throw new Error(text || 'Не удалось загрузить меню');
            }
        }

        const data = await response.json();
        const restaurant = data.restaurant;
        const menuItems = data.menu_items;

        if (!restaurant) {
            throw new Error('Информация о ресторане не найдена');
        }

        const restaurantName = document.getElementById('restaurant-name');
        const restaurantDescription = document.getElementById('restaurant-description');
        const restaurantImage = document.getElementById('restaurant-image');
        const restaurantAddress = document.getElementById('restaurant-address');

        if (restaurantName) {
            restaurantName.textContent = restaurant.name;
        }
        if (restaurantDescription) {
            restaurantDescription.textContent = `Испытайте вкус ${restaurant.cuisine_type && restaurant.cuisine_type.String ? restaurant.cuisine_type.String : 'традиционной кухни'} с нашими аутентичными блюдами.`;
        }
        if (restaurantImage) {
            restaurantImage.src = '/static/images/restaurant-placeholder.jpg';
        }
        if (restaurantAddress) {
            restaurantAddress.textContent = restaurant.address && restaurant.address.String ? restaurant.address.String : 'Адрес не указан';
        }

        const menuGrid = document.getElementById('menu-grid');
        if (menuGrid) {
            menuGrid.innerHTML = '';
            if (!menuItems || menuItems.length === 0) {
                menuGrid.innerHTML = '<p>Меню пусто.</p>';
                return;
            }

            menuItems.forEach(item => {
                const menuItemId = item.id;
                if (!menuItemId || isNaN(menuItemId)) {
                    console.error('Неверный ID блюда:', item);
                    return;
                }

                const menuItem = document.createElement('div');
                menuItem.className = 'menu-item';
                menuItem.innerHTML = `
                    <img src="${item.image_url && item.image_url.String ? item.image_url.String : '/static/images/food-placeholder.jpg'}" alt="${item.name}">
                    <div class="menu-item-details">
                        <h3>${item.name}</h3>
                        <p class="price">${item.price.toFixed(2)} ₽</p>
                    </div>
                `;
                menuItem.addEventListener('click', () => {
                    window.location.href = `/product-details/${menuItemId}`;
                });
                menuGrid.appendChild(menuItem);
            });
        }

        const reviewsSection = document.getElementById('reviews');
        if (reviewsSection) {
            reviewsSection.innerHTML = `
                <div class="review">
                    <p>"Удивительная еда и фантастическая атмосфера. Обязательно попробуйте блюда!"</p>
                    <p class="review-author">— София Л.</p>
                </div>
                <div class="review">
                    <p>"Отличный сервис и вкусная еда. Рекомендую!"</p>
                    <p class="review-author">— Михаил Б.</p>
                </div>
            `;
        }
    } catch (error) {
        console.error('Ошибка при загрузке меню:', error);
        displayError('menu-grid', error.message);
    }
});
