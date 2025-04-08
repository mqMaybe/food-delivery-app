document.addEventListener('DOMContentLoaded', async () => {
    const session = await checkAuth('customer');
    if (!session) return;

    const loadMyOrdersBtn = document.getElementById('loadMyOrdersBtn');
    const myOrdersList = document.getElementById('myOrdersList');

    if (!loadMyOrdersBtn || !myOrdersList) {
        console.error('Required elements not found');
        return;
    }

    loadMyOrdersBtn.addEventListener('click', async () => {
        try {
            const response = await fetch('/api/orders');
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Не удалось загрузить заказы');
            }

            const orders = await response.json();
            if (orders.length === 0) {
                myOrdersList.innerHTML = '<p>У вас нет заказов.</p>';
                return;
            }

            myOrdersList.innerHTML = '';
            orders.forEach(order => {
                const div = document.createElement('div');
                let itemsHtml = order.items.map(item => `
                    <li>${item.menu_name} - ${item.menu_price} ₽ x ${item.quantity}</li>
                `).join('');
                div.innerHTML = `
                    <p>Заказ #${order.id} | Статус: ${order.status}</p>
                    <p>Адрес доставки: ${order.delivery_address}</p>
                    <p>Итого: ${order.total_price} ₽</p>
                    <ul>${itemsHtml}</ul>
                `;
                myOrdersList.appendChild(div);
            });
        } catch (error) {
            console.error('Failed to load orders:', error);
            displayError('myOrdersList', error.message);
        }
    });
});