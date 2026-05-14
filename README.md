Система загрузки и агрегации данных из Wildberries, Ozon и МойСклад.

## Стек

- Node.js (ES modules)
- Express.js
- PostgreSQL
- Axios
- Luxon

## Возможности

- Загрузка заказов WB / Ozon (FBO, FBS)
- Синхронизация остатков и карточек товаров
- Интеграция с МойСклад (остатки по складам)
- REST API для фронтенда
- Логирование синхронизаций

## Установка

npm install

## Запуск

## разработка с nodemon
npm run dev

## production
npm start

## Мок без БД
npm run mock

## API эндпоинты

/api/wb/orders – заказы Wildberries
/api/wb/remains – остатки WB
/api/wb/cards – карточки товаров
/api/ozon/orders – заказы Ozon
/api/ozon/remains – остатки Ozon
/api/moysklad/stocks – остатки МойСклад
/api/dashboard/stats – общая статистика
/api/charts/orders-daily – динамика заказов
/api/sync/logs – логи синхронизации

## Переменные окружения

Создайте .env файл:
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=wb_orders
WB_API_TOKEN=ваш_токен
OZON_CLIENT_ID=client_id
OZON_API_KEY=api_key
MS_TOKEN=токен_мoйсклад
PORT=3000
