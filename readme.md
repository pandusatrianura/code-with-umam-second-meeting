<div align="center">
<a href="https://github.com/Sumonta056/FixHub-Issue-Tracker-Website" target="blank">
<img src="./assets/chatBot.png" width="90" alt="Logo" />
</a>

<h2> Project Name : Kasir API </h2>


[![Postman](https://img.shields.io/badge/Postman-FF6C37?logo=postman&logoColor=white)](#)
[![Swagger](https://img.shields.io/badge/Swagger-85EA2D?logo=swagger&logoColor=173647)](#)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-6BA539?logo=openapiinitiative&logoColor=white)](#)
[![IntelliJ IDEA](https://img.shields.io/badge/IntelliJIDEA-000000.svg?logo=intellij-idea&logoColor=white)](#)
[![Supabase](https://img.shields.io/badge/Supabase-3FCF8E?logo=supabase&logoColor=fff)](#)
[![Go](https://img.shields.io/badge/Go-00ADD8?logo=Go&logoColor=white&style=for-the-badge)](#)

<img src ="./assets/web.webp" width="80%">

</div>

## 🎯 Task 

- [x] Session 1: https://docs.kodingworks.io/s/17137b9a-ed7a-4950-ba9e-eb11299531c2#h-%F0%9F%8E%AF-tugas
- [x] Session 2: https://docs.kodingworks.io/s/a378a9fe-c0e0-4fa1-a896-43ae347a7b61

## 💡 Overview

### Task Session 1
1. Implementasikan CRUD Kategori pada project API kalian.

### Task Session 2
1. Pindah categories temen-temen ke layered architecture.
2. **Challange (Optional):** Explore Join, tambah category_id ke table products, setiap product mempunyai kategory, dan ketika Get Detail return category.name dari product.

```json
{
  "code": "1000",
  "message": "Product retrieved successfully",
  "data": {
    "id": 1,
    "name": "Bebelac",
    "price": 10000,
    "stock": 100,
    "category_name": "Susu",
    "created_at": "2026-01-29T04:27:44.690013+07:00",
    "updated_at": "2026-01-29T04:27:44.690013+07:00"
  }
}
```

## ✨ Model

### Category
- **ID**
- **Name**
- **Description**
- **Created At**
- **Updated At**

### Product
- **ID**
- **Name**
- **Price**
- **Stock**
- **Category ID**
- **Created At**
- **Updated At**

## 📖 API Endpoints

The application provides several API endpoints for the functionalities mentioned above. Below are some key endpoints:

### Category
- **Ambil semua kategori**: `GET /categories`
- **Tambah satu kategori**: `POST /categories`
- **Update satu kategori**: `PUT /categories/{id}`
- **Ambil detail satu kategori**: `GET /categories/{id}`
- **Hapus satu kategori**: `DELETE /categories/{id}`

### Product
- **Ambil semua produk**: `GET /products`
- **Tambah satu produk**: `POST /products`
- **Update satu produk**: `PUT /products/{id}`
- **Ambil detail satu kategori**: `GET /products/{id}`
- **Hapus satu kategori**: `DELETE /products/{id}`

## 🛠️ Installation

1. **Clone the Repository**:
   ```bash
   git clone <repository-url>
   ```

2. **Install Dependencies**:
   ```bash
   cd <project-directory>
   go mod tidy
   ```

3. **Run the Application**:
   ```bash
   go run main.go 
   ```

## 📦 Access the API Documentation
    
1. Use tools like Postman or cURL to interact with the API endpoints.
    
   ```bash
   postman collection: docs/kasir-api.postman_collection.json
   postman environment: docs/kasir-api.postman_environment.json
   ```
2. Use API Documentation to interact with the API endpoints.
    
    ```bash
    swagger: http://{{url}}/api/docs
   ```
   Note: 
   - Replace `{{url}}` with the URL of your deployed API (see 📖 Hosted API).
   - This docs are generated using [swaggo](https://github.com/swaggo/swag).
   - You can also generate the docs locally using `swag init`.
   - This docs run under HTTP protocol.

## 📃 List of API Endpoints

### Category

1. Health Check Endpoint:
   ```bash
   curl --location '{{url}}/api/categories/health'
   ```
2. Display All Categories Endpoint:
   ```bash
   curl --location '{{url}}/api/categories'
   ```
3. Display Category By ID Endpoint:
   ```bash
   curl --location '{{url}}/api/categories/6'
   ```
4. Create New Category Endpoint:
   ```bash
   curl --location '{{url}}/api/categories' \
   --header 'Content-Type: application/json' \
   --data '{
   "name": "Susu",
   "description": "Kategori Susu"
   }'
   ```
5. Update Existing Category Endpoint:
   ```bash
   curl --location --request PUT '{{url}}/api/categories/9' \
   --header 'Content-Type: application/json' \
   --data '{
   "name": "Minuman",
   "description": "Kategori Minuman"
   }'
   ```
6. Delete Existing Category Endpoint:
   ```bash
   curl --location --request DELETE '{{url}}/api/categories/9'
   ```
### Product

1. Health Check Endpoint:
   ```bash
   curl --location '{{url}}/api/products/health'
   ```
2. Display All Products Endpoint:
   ```bash
   curl --location '{{url}}/api/products'
   ```
3. Display Product By ID Endpoint:
   ```bash
   curl --location '{{url}}/api/products/6'
   ```
4. Create New Product Endpoint:
   ```bash
   curl --location '{{url}}/api/v1/products' \
   --header 'Content-Type: application/json' \
   --data '{
    "name": "Bebelac",
    "price": 10000,
    "stock": 100,
    "category_id": 2
   }'
   ```
5. Update Existing Product Endpoint:
   ```bash
   curl --location --request PUT '{{url}}/api/products/9' \
   --header 'Content-Type: application/json' \
   --data '{
    "name": "Bebelac",
    "price": 10000,
    "stock": 10,
    "category_id": 2
   }'
   ```
6. Delete Existing Product Endpoint:
   ```bash
   curl --location --request DELETE '{{url}}/api/products/9'
   ```

## 📖 Hosted API

Please change the `{{url}}` to the URL of your deployed API.

   Localhost:
   ```bash
   http://localhost:8000
   ```
   Public:
   ```bash
   https://kasir.pandusatrianura.cloud
   ```

## 📜 License

Distributed under the MIT License. See [License](/LICENSE) for more information.