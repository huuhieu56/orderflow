# OrderFlow — Project Overview

## 1. Giới thiệu

OrderFlow là một hệ thống đặt hàng theo kiến trúc microservices, được xây dựng chủ yếu bằng Golang.

Mục tiêu của dự án là mô phỏng một hệ thống backend hiện đại có đầy đủ các thành phần thường xuất hiện trong môi trường production, bao gồm:

* Authentication và authorization
* Database riêng cho từng service
* Redis cache
* Kafka message broker
* Event-driven architecture
* Docker
* Kubernetes
* Logging, metrics và distributed tracing
* Transactional Outbox Pattern
* Idempotent Consumer
* Retry và Dead Letter Queue

OrderFlow không tập trung vào nghiệp vụ thương mại điện tử phức tạp. Dự án chỉ giữ lại các chức năng cần thiết để người phát triển có thể học và thực hành cách xây dựng một hệ thống phân tán hoàn chỉnh.

Hệ thống không có thanh toán thật, giao hàng, mã giảm giá, tìm kiếm sản phẩm hoặc frontend phức tạp.

---

# 2. Mục tiêu dự án

Sau khi hoàn thành OrderFlow, người phát triển cần hiểu và thực hành được:

* Cách tổ chức một backend service bằng Go
* Cách chia một hệ thống thành nhiều microservice
* Cách mỗi service sở hữu dữ liệu riêng
* Cách xác thực người dùng bằng JWT
* Cách sử dụng Redis để cache và rate limiting
* Cách producer gửi event lên Kafka
* Cách consumer nhận và xử lý event
* Cách Kafka partition, offset và consumer group hoạt động
* Cách tránh mất event bằng Transactional Outbox Pattern
* Cách tránh xử lý trùng message bằng Idempotent Consumer
* Cách xử lý retry và dead-letter event
* Cách container hóa hệ thống bằng Docker
* Cách deploy hệ thống lên Kubernetes
* Cách theo dõi hệ thống bằng logging, metrics và tracing

---

# 3. Phạm vi nghiệp vụ

OrderFlow mô phỏng một hệ thống đặt hàng đơn giản.

Người dùng có thể:

* Đăng ký tài khoản
* Đăng nhập
* Xem thông tin cá nhân
* Xem danh sách sản phẩm
* Xem chi tiết sản phẩm
* Tạo đơn hàng
* Xem danh sách đơn hàng của mình
* Xem chi tiết đơn hàng
* Hủy đơn hàng
* Xem thông báo liên quan đến đơn hàng

Quản trị viên có thể:

* Tạo sản phẩm
* Cập nhật sản phẩm
* Thay đổi trạng thái sản phẩm
* Xem danh sách sản phẩm

Khi một đơn hàng được tạo hoặc hủy, hệ thống phát một event lên Kafka.

Notification Service nhận event này và tạo thông báo cho người dùng.

---

# 4. Kiến trúc tổng quan

Hệ thống gồm bốn service chính:

* Auth Service
* Product Service
* Order Service
* Notification Service

Các thành phần hạ tầng:

* PostgreSQL
* Redis
* Kafka
* API Gateway hoặc Kubernetes Ingress
* Prometheus
* Grafana
* OpenTelemetry Collector
* Docker
* Kubernetes

Luồng tổng quát:

Client gửi request đến API Gateway hoặc Ingress.

Gateway định tuyến request đến service tương ứng.

Mỗi service xử lý nghiệp vụ và truy cập database của riêng mình.

Các service không truy cập trực tiếp database của nhau.

Order Service giao tiếp bất đồng bộ với Notification Service thông qua Kafka.

---

# 5. Nguyên tắc thiết kế

## 5.1. Database per Service

Mỗi service sở hữu dữ liệu riêng.

Auth Service chỉ quản lý dữ liệu người dùng.

Product Service chỉ quản lý dữ liệu sản phẩm.

Order Service chỉ quản lý dữ liệu đơn hàng.

Notification Service chỉ quản lý dữ liệu thông báo.

Một service không được đọc hoặc ghi trực tiếp database của service khác.

Khi cần dữ liệu từ service khác, hệ thống sử dụng một trong hai cách:

* Gọi API đồng bộ
* Nhận dữ liệu thông qua event bất đồng bộ

Nguyên tắc này giúp giảm phụ thuộc giữa các service và tránh việc nhiều service cùng thay đổi một schema database.

## 5.2. Stateless Service

Các Go service được thiết kế theo hướng stateless.

Service không lưu trạng thái người dùng trong memory cục bộ.

Những dữ liệu cần chia sẻ giữa nhiều instance được lưu trong:

* PostgreSQL
* Redis
* Kafka

Nhờ đó, nhiều instance của cùng một service có thể chạy song song trong Kubernetes.

## 5.3. Event-driven Architecture

Những tác vụ không cần trả kết quả ngay cho client được xử lý bất đồng bộ qua Kafka.

Ví dụ:

* Tạo thông báo sau khi đơn hàng được tạo
* Ghi nhận dữ liệu phân tích
* Gửi email
* Đồng bộ dữ liệu sang hệ thống khác

Order Service chỉ chịu trách nhiệm hoàn tất nghiệp vụ tạo đơn hàng và phát event.

Notification Service chịu trách nhiệm xử lý event và tạo thông báo.

## 5.4. Failure Isolation

Mỗi service có thể bị lỗi độc lập.

Nếu Notification Service tạm thời dừng hoạt động, Order Service vẫn có thể tạo đơn hàng.

Event vẫn được giữ trong Kafka và sẽ được Notification Service xử lý sau khi service hoạt động trở lại.

---

# 6. Các service chính

## 6.1. Auth Service

Auth Service chịu trách nhiệm quản lý danh tính người dùng.

### Trách nhiệm

* Đăng ký tài khoản
* Đăng nhập
* Mã hóa mật khẩu
* Phát hành access token
* Phát hành refresh token
* Làm mới access token
* Đăng xuất
* Kiểm tra thông tin người dùng
* Quản lý role cơ bản
* Giới hạn số lần đăng nhập thất bại

### Dữ liệu chính

User bao gồm:

* User ID
* Email
* Password hash
* Role
* Account status
* Created time
* Updated time

### Role

Dự án chỉ cần hai role:

* User
* Admin

User có thể xem sản phẩm và quản lý đơn hàng của mình.

Admin có thể quản lý sản phẩm.

### Redis trong Auth Service

Redis được dùng để:

* Lưu refresh token
* Thu hồi refresh token
* Rate limit login
* Lưu thông tin đăng nhập thất bại trong thời gian ngắn
* Hỗ trợ token blacklist nếu cần

### Kết quả đầu ra

Sau khi đăng nhập thành công, người dùng nhận:

* Access token
* Refresh token
* Thời gian hết hạn
* Thông tin cơ bản của tài khoản

---

## 6.2. Product Service

Product Service chịu trách nhiệm quản lý sản phẩm.

### Trách nhiệm

* Tạo sản phẩm
* Xem danh sách sản phẩm
* Xem chi tiết sản phẩm
* Cập nhật sản phẩm
* Bật hoặc tắt sản phẩm
* Quản lý số lượng tồn kho đơn giản

### Dữ liệu chính

Product bao gồm:

* Product ID
* Product name
* Description
* Price
* Stock quantity
* Product status
* Created time
* Updated time

### Product status

Sản phẩm có thể có các trạng thái:

* Active
* Inactive

Chỉ những sản phẩm Active mới được người dùng bình thường nhìn thấy.

### Redis trong Product Service

Redis được sử dụng để cache:

* Chi tiết sản phẩm
* Danh sách sản phẩm
* Những truy vấn sản phẩm phổ biến

Khi sản phẩm được cập nhật, Product Service phải xóa hoặc cập nhật cache liên quan.

### Cache strategy

Dự án sử dụng Cache-Aside Pattern.

Quy trình đọc:

* Service kiểm tra Redis
* Nếu có dữ liệu, trả dữ liệu từ cache
* Nếu không có, đọc PostgreSQL
* Lưu dữ liệu vào Redis
* Trả kết quả cho client

Quy trình cập nhật:

* Cập nhật PostgreSQL
* Xóa cache liên quan
* Request tiếp theo sẽ tạo lại cache

---

## 6.3. Order Service

Order Service là service trung tâm của hệ thống.

### Trách nhiệm

* Tạo đơn hàng
* Xem danh sách đơn hàng của người dùng
* Xem chi tiết đơn hàng
* Hủy đơn hàng
* Tính tổng giá trị đơn hàng
* Lưu snapshot sản phẩm tại thời điểm đặt hàng
* Phát event liên quan đến đơn hàng
* Quản lý Transactional Outbox

### Dữ liệu chính

Order bao gồm:

* Order ID
* User ID
* Order status
* Total amount
* Created time
* Updated time

Order Item bao gồm:

* Order Item ID
* Order ID
* Product ID
* Product name
* Unit price
* Quantity
* Subtotal

### Order status

Phiên bản đầu sử dụng các trạng thái:

* Pending
* Confirmed
* Cancelled

Luồng đơn giản:

* Người dùng tạo đơn
* Order được tạo với trạng thái Pending
* Sau khi validation hoàn tất, order chuyển thành Confirmed
* Người dùng có thể hủy order nếu order chưa bị hủy trước đó

### Product snapshot

Order Service không nên phụ thuộc hoàn toàn vào dữ liệu hiện tại của Product Service.

Khi tạo đơn hàng, Order Service lưu lại:

* Tên sản phẩm
* Giá sản phẩm
* Số lượng
* Product ID

Nếu giá sản phẩm thay đổi sau đó, đơn hàng cũ vẫn giữ nguyên giá tại thời điểm đặt hàng.

### Giao tiếp với Product Service

Khi tạo đơn hàng, Order Service cần xác nhận:

* Sản phẩm tồn tại
* Sản phẩm đang Active
* Số lượng hợp lệ
* Giá hiện tại
* Tồn kho đủ

Ở phiên bản đầu, Order Service gọi trực tiếp Product Service qua HTTP hoặc gRPC.

Để giảm độ phức tạp, Product Service có thể chỉ kiểm tra dữ liệu và chưa cần triển khai cơ chế reservation tồn kho.

### Event được phát

Order Service phát các event:

* order.created
* order.confirmed
* order.cancelled

MVP chỉ cần:

* order.created
* order.cancelled

---

## 6.4. Notification Service

Notification Service xử lý thông báo cho người dùng.

### Trách nhiệm

* Consume event từ Kafka
* Tạo notification
* Lưu notification vào database
* Cung cấp danh sách notification cho người dùng
* Đánh dấu notification đã đọc
* Chống xử lý trùng event
* Retry khi xử lý thất bại
* Đẩy event lỗi vào Dead Letter Topic

### Dữ liệu chính

Notification bao gồm:

* Notification ID
* User ID
* Notification type
* Title
* Content
* Read status
* Created time

### Các loại thông báo

* Order created
* Order confirmed
* Order cancelled

### Email

Phiên bản đầu không cần gửi email thật.

Notification Service chỉ cần:

* Tạo notification trong database
* Ghi log mô phỏng gửi email

Gửi email thật có thể được thêm ở phiên bản sau.

---

# 7. Authentication và Authorization

## 7.1. Access Token

Access token là JWT có thời gian sống ngắn.

Token chứa các thông tin cơ bản:

* User ID
* Email
* Role
* Expiration time
* Token ID

Client gửi token trong header của request.

Mỗi service tự xác minh JWT.

## 7.2. Refresh Token

Refresh token có thời gian sống dài hơn.

Refresh token được lưu hoặc quản lý thông qua Redis.

Khi access token hết hạn, client sử dụng refresh token để lấy access token mới.

Khi logout, refresh token bị xóa hoặc thu hồi.

## 7.3. Authorization

Authorization dựa trên role và quyền sở hữu tài nguyên.

Ví dụ:

* Chỉ admin được tạo sản phẩm
* User chỉ xem được order của chính mình
* User chỉ xem được notification của chính mình
* Admin không mặc định được sửa dữ liệu auth của user nếu chưa có chức năng quản trị rõ ràng

## 7.4. Service-to-service Authentication

Ở phiên bản đầu, các service nội bộ có thể giao tiếp trong private network của Kubernetes.

Phiên bản nâng cao có thể bổ sung:

* Internal API key
* Service JWT
* Mutual TLS
* Service mesh

Những phần này không bắt buộc trong MVP.

---

# 8. Redis

Redis được dùng trong hai nhóm chức năng.

## 8.1. Authentication

* Lưu refresh token
* Token revocation
* Login rate limiting
* Temporary login state

## 8.2. Product cache

* Cache product detail
* Cache product list
* Giảm số lần truy cập PostgreSQL
* Cải thiện response time

## 8.3. Quy tắc sử dụng Redis

Redis không phải nguồn dữ liệu chính.

PostgreSQL vẫn là source of truth.

Nếu Redis mất dữ liệu, hệ thống vẫn phải hoạt động bằng cách đọc lại từ PostgreSQL.

Các key Redis phải có expiration time phù hợp.

Không nên giữ cache vĩnh viễn.

---

# 9. Kafka

Kafka là message broker chính của OrderFlow.

## 9.1. Vai trò

Kafka được dùng để:

* Truyền event giữa các service
* Giảm coupling
* Cho phép xử lý bất đồng bộ
* Lưu event trong một khoảng thời gian
* Hỗ trợ nhiều consumer độc lập
* Scale consumer theo partition

## 9.2. Topic chính

Topic ban đầu:

* order.events

Topic này chứa nhiều loại event liên quan đến order.

Event type được định nghĩa trong message payload.

## 9.3. Message key

Order ID được dùng làm Kafka message key.

Điều này giúp các event của cùng một order đi vào cùng partition.

Nhờ đó, thứ tự xử lý event của một order được giữ ổn định trong cùng partition.

## 9.4. Partition

MVP có thể bắt đầu với một partition.

Sau khi hệ thống chạy ổn, topic được tăng lên ba partition để thực hành:

* Consumer group
* Load balancing
* Parallel processing
* Ordering theo key

## 9.5. Consumer group

Notification Service sử dụng một consumer group riêng.

Nếu có nhiều instance Notification Service, Kafka phân phối partition giữa các instance trong cùng group.

Một event chỉ được một instance trong consumer group xử lý.

Nếu sau này có Analytics Service, service này sử dụng consumer group khác.

Notification Service và Analytics Service đều nhận được event vì chúng thuộc các consumer group khác nhau.

## 9.6. Offset

Consumer chỉ commit offset sau khi xử lý nghiệp vụ thành công.

Quy trình:

* Nhận message
* Validate event
* Kiểm tra duplicate
* Lưu notification
* Ghi nhận event đã xử lý
* Commit database transaction
* Commit Kafka offset

Nếu consumer lỗi trước khi commit offset, message có thể được gửi lại.

---

# 10. Event contract

Mỗi event cần có cấu trúc thống nhất.

Thông tin metadata gồm:

* Event ID
* Event type
* Event version
* Occurred time
* Producer
* Correlation ID
* Trace ID

Payload chứa dữ liệu nghiệp vụ.

Ví dụ event order.created chứa:

* Order ID
* User ID
* Total amount
* Order status
* Created time

## Event versioning

Event phải có version.

Khi schema event thay đổi, consumer có thể xác định version để xử lý tương thích.

Không nên sửa event cũ theo cách phá vỡ consumer đang chạy.

---

# 11. Transactional Outbox Pattern

## 11.1. Vấn đề

Nếu Order Service thực hiện hai thao tác riêng biệt:

* Lưu order vào PostgreSQL
* Publish event lên Kafka

Có thể xảy ra trường hợp:

* Order đã được lưu
* Publish event thất bại

Khi đó Notification Service không biết order đã được tạo.

## 11.2. Giải pháp

Order Service sử dụng Outbox Pattern.

Trong cùng một database transaction:

* Lưu order
* Lưu order item
* Lưu outbox event

Sau khi transaction commit, một background worker đọc các outbox event chưa publish.

Worker gửi event lên Kafka.

Khi publish thành công, outbox event được đánh dấu là published.

## 11.3. Outbox status

Outbox event có thể có các trạng thái:

* Pending
* Processing
* Published
* Failed

## 11.4. Retry

Nếu publish thất bại:

* Tăng retry count
* Ghi lại lỗi gần nhất
* Chờ một khoảng thời gian
* Thử lại

Sau quá nhiều lần thất bại, event có thể được chuyển sang trạng thái Failed để kiểm tra thủ công.

---

# 12. Idempotent Consumer

## 12.1. Vấn đề

Kafka consumer có thể nhận cùng một event nhiều lần.

Ví dụ:

* Consumer lưu notification thành công
* Consumer crash trước khi commit offset
* Kafka gửi lại message

Nếu không có idempotency, Notification Service tạo hai notification giống nhau.

## 12.2. Giải pháp

Notification Service lưu danh sách event đã xử lý.

Mỗi event có Event ID duy nhất.

Trong một database transaction:

* Kiểm tra Event ID
* Nếu đã tồn tại, bỏ qua
* Nếu chưa tồn tại, tạo notification
* Lưu Event ID vào processed events
* Commit transaction

Sau đó mới commit Kafka offset.

Nhờ đó, cùng một event có thể được gửi nhiều lần nhưng kết quả nghiệp vụ chỉ được tạo một lần.

---

# 13. Retry và Dead Letter Topic

## 13.1. Lỗi có thể retry

Một số lỗi chỉ mang tính tạm thời:

* Database connection lỗi
* Network timeout
* Service phụ thuộc chưa sẵn sàng
* Kafka broker tạm thời không phản hồi

Các event này có thể được retry.

## 13.2. Lỗi không nên retry liên tục

Một số lỗi không thể tự khắc phục:

* Event payload sai format
* Thiếu field bắt buộc
* Event version không được hỗ trợ
* Dữ liệu nghiệp vụ không hợp lệ

Những event này nên được đưa vào Dead Letter Topic.

## 13.3. Topic đề xuất

* order.events
* order.events.retry
* order.events.dlq

MVP chưa cần triển khai retry topic ngay.

Có thể bổ sung sau khi luồng producer-consumer cơ bản đã ổn định.

---

# 14. API Gateway và Ingress

## 14.1. MVP

Phiên bản đầu không cần viết API Gateway riêng bằng Go.

Kubernetes Ingress hoặc reverse proxy sẽ định tuyến request.

Các route chính:

* Auth route đến Auth Service
* Product route đến Product Service
* Order route đến Order Service
* Notification route đến Notification Service

## 14.2. Phiên bản nâng cao

Sau này có thể xây Go API Gateway để xử lý:

* Authentication middleware
* Rate limiting
* Request ID
* Centralized logging
* Reverse proxy
* Timeout
* Retry
* Circuit breaker
* Response aggregation

Gateway không nên chứa business logic.

---

# 15. Database

## 15.1. PostgreSQL

PostgreSQL là database chính cho tất cả service.

Mỗi service có database hoặc schema riêng.

Trong local development, có thể dùng một PostgreSQL instance với nhiều database.

Trong production, tùy quy mô có thể tách thành nhiều instance khác nhau.

## 15.2. Database ownership

Auth Service sở hữu:

* Users
* Refresh token metadata nếu không lưu hoàn toàn trong Redis

Product Service sở hữu:

* Products

Order Service sở hữu:

* Orders
* Order items
* Outbox events

Notification Service sở hữu:

* Notifications
* Processed events

## 15.3. Migration

Mỗi service tự quản lý migration của mình.

Migration được chạy độc lập theo từng service.

Không có một migration chung thay đổi database của nhiều service cùng lúc.

---

# 16. Giao tiếp giữa các service

## 16.1. Giao tiếp đồng bộ

Được dùng khi service cần kết quả ngay.

Ví dụ:

Order Service gọi Product Service để lấy:

* Product name
* Product price
* Product status
* Available stock

Giao tiếp có thể dùng HTTP trong phiên bản đầu.

gRPC có thể được thêm ở phiên bản sau.

## 16.2. Giao tiếp bất đồng bộ

Được dùng khi service không cần kết quả ngay.

Ví dụ:

Order Service phát order.created.

Notification Service nhận event và tạo notification.

## 16.3. Timeout

Mọi lời gọi service-to-service phải có timeout.

Không được để một request chờ vô hạn.

## 16.4. Retry

Retry chỉ nên dùng cho lỗi tạm thời.

Retry phải có giới hạn.

Không retry vô hạn vì có thể gây tăng tải lên hệ thống.

## 16.5. Circuit breaker

Circuit breaker là phần nâng cao.

Có thể thêm sau khi hệ thống cơ bản hoàn thành.

Mục tiêu là ngăn một service tiếp tục gọi liên tục đến service đang lỗi.

---

# 17. Docker

Mỗi Go service có Docker image riêng.

Docker Compose được dùng cho local development.

Hệ thống local gồm:

* Auth Service
* Product Service
* Order Service
* Notification Service
* PostgreSQL
* Redis
* Kafka
* Monitoring components nếu cần

Mục tiêu của Docker Compose:

* Khởi động toàn bộ hệ thống bằng một lệnh
* Tạo network nội bộ
* Cấu hình environment variables
* Hỗ trợ phát triển và kiểm thử local
* Dễ dàng reset hệ thống

---

# 18. Kubernetes

Kubernetes được dùng để deploy hệ thống sau khi Docker Compose chạy ổn định.

## 18.1. Resource cho mỗi service

Mỗi Go service có:

* Deployment
* Service
* ConfigMap
* Secret
* Liveness probe
* Readiness probe
* Resource requests
* Resource limits

## 18.2. Ingress

Ingress định tuyến request từ bên ngoài đến các service.

## 18.3. Scaling

Các stateless Go service có thể scale thành nhiều replica.

Notification Service có thể scale dựa trên:

* CPU
* Memory
* Kafka consumer lag

## 18.4. Stateful component

PostgreSQL, Redis và Kafka là stateful component.

Trong môi trường học tập, có thể chạy chúng trong Kubernetes.

Trong môi trường production thực tế, thường ưu tiên managed service hoặc hệ thống được vận hành chuyên biệt.

## 18.5. Local Kubernetes

Dự án có thể sử dụng:

* Kind
* Minikube

Kind phù hợp để tạo cluster local nhẹ và dễ reset.

---

# 19. Health check

Mỗi service phải có hai loại health check.

## Liveness

Xác định process còn hoạt động hay không.

Nếu liveness thất bại nhiều lần, Kubernetes restart container.

## Readiness

Xác định service có sẵn sàng nhận traffic hay không.

Readiness có thể kiểm tra:

* Database connection
* Redis connection nếu Redis là dependency bắt buộc
* Kafka connection nếu service cần Kafka để hoạt động

Không nên kiểm tra quá nhiều dependency trong liveness.

---

# 20. Logging

Mỗi service sử dụng structured logging.

Log nên chứa:

* Timestamp
* Log level
* Service name
* Environment
* Request ID
* User ID nếu phù hợp
* Trace ID
* Error message
* Event ID
* Order ID
* Kafka topic
* Kafka partition
* Kafka offset

Không ghi vào log:

* Password
* Access token
* Refresh token
* Dữ liệu nhạy cảm

Log được ghi dưới dạng JSON để dễ thu thập và tìm kiếm.

---

# 21. Metrics

Prometheus được dùng để thu thập metrics.

Các metrics cơ bản:

* Request count
* Request latency
* Error count
* Active request
* Database query duration
* Redis cache hit
* Redis cache miss
* Kafka event published
* Kafka publish failure
* Kafka event consumed
* Kafka consumer error
* Outbox pending count
* Outbox failed count
* Notification created count
* Kafka consumer lag

Grafana được dùng để tạo dashboard.

---

# 22. Distributed tracing

OpenTelemetry được dùng để theo dõi request qua nhiều service.

Một trace có thể mô tả luồng:

* Client gọi Order Service
* Order Service gọi Product Service
* Order Service ghi PostgreSQL
* Outbox Worker publish event
* Notification Service consume event
* Notification Service ghi PostgreSQL

Trace context cần được truyền qua:

* HTTP headers
* Kafka message headers

Tracing giúp xác định service nào chậm hoặc lỗi trong toàn bộ luồng.

---

# 23. Security

Các nguyên tắc bảo mật chính:

* Password được hash
* Không lưu password dạng plain text
* JWT có thời hạn
* Refresh token có thể thu hồi
* Secret không nằm trực tiếp trong source code
* Kubernetes Secret quản lý dữ liệu nhạy cảm
* Input phải được validate
* User không được truy cập tài nguyên của user khác
* Admin endpoint phải kiểm tra role
* API có rate limiting
* Log không chứa token hoặc password
* Database user chỉ có quyền cần thiết
* Container chạy bằng non-root user nếu có thể

---

# 24. Testing strategy

## Unit test

Kiểm tra business logic độc lập.

Ví dụ:

* Tính tổng giá trị order
* Kiểm tra order status
* Kiểm tra quyền truy cập
* Validate input
* Xử lý duplicate event

## Integration test

Kiểm tra service với dependency thật.

Ví dụ:

* Repository với PostgreSQL
* Cache với Redis
* Producer và consumer với Kafka
* Outbox Worker

## API test

Kiểm tra các API chính.

Ví dụ:

* Register
* Login
* Create product
* Create order
* Get notifications

## End-to-end test

Kiểm tra toàn bộ luồng:

* User đăng ký
* User đăng nhập
* Admin tạo product
* User tạo order
* Order được lưu
* Event được publish
* Notification Service consume
* Notification xuất hiện trong tài khoản user

---

# 25. Error handling

Mỗi service cần có format lỗi thống nhất.

Error response bao gồm:

* Error code
* Error message
* Request ID
* Validation details nếu có

Các nhóm lỗi:

* Validation error
* Authentication error
* Authorization error
* Resource not found
* Conflict
* Rate limit exceeded
* Internal server error
* Dependency unavailable

Không trả chi tiết lỗi database hoặc stack trace cho client.

---

# 26. Configuration

Cấu hình được quản lý bằng environment variables.

Các nhóm cấu hình:

* HTTP server
* Database
* Redis
* Kafka
* JWT
* Logging
* Metrics
* Tracing
* Cache TTL
* Retry policy

Local development có thể dùng file environment.

Kubernetes sử dụng ConfigMap và Secret.

---

# 27. Development phases

## Phase 1 — Core Monolith

Xây một ứng dụng Go duy nhất gồm:

* Auth
* Product
* Order
* Notification

Sử dụng PostgreSQL.

Mục tiêu là hoàn thành business flow trước.

## Phase 2 — Authentication

Hoàn thiện:

* Register
* Login
* JWT
* Refresh token
* Role
* Resource ownership

## Phase 3 — Redis

Thêm:

* Product cache
* Login rate limiting
* Refresh token management

## Phase 4 — Tách Notification Service

Notification là service đầu tiên được tách vì nó có thể giao tiếp hoàn toàn qua Kafka.

Order Service publish order.created.

Notification Service consume event.

## Phase 5 — Tách các service còn lại

Tách lần lượt:

* Product Service
* Auth Service
* Order Service

Mỗi service có database riêng.

## Phase 6 — Kafka reliability

Thêm:

* Consumer group
* Manual offset commit
* Transactional Outbox
* Idempotent Consumer
* Retry
* Dead Letter Topic

## Phase 7 — Docker Compose

Container hóa toàn bộ hệ thống.

Chạy tất cả service và infrastructure bằng Docker Compose.

## Phase 8 — Kubernetes

Deploy lên Kind hoặc Minikube.

Thêm:

* Deployment
* Service
* Ingress
* ConfigMap
* Secret
* Health probe
* Resource limit

## Phase 9 — Observability

Thêm:

* Structured logging
* Prometheus
* Grafana
* OpenTelemetry
* Kafka consumer lag monitoring

## Phase 10 — Hardening

Thêm:

* Timeout
* Retry policy
* Graceful shutdown
* Better validation
* Security review
* Load testing
* Failure testing

---

# 28. MVP Definition

MVP được xem là hoàn thành khi hệ thống thực hiện được luồng sau:

* Người dùng đăng ký
* Người dùng đăng nhập
* Người dùng nhận JWT
* Admin tạo sản phẩm
* Người dùng xem sản phẩm
* Product được cache trong Redis
* Người dùng tạo đơn hàng
* Order được lưu vào PostgreSQL
* Order Service tạo event
* Event được publish lên Kafka
* Notification Service consume event
* Notification được lưu
* Người dùng xem được notification
* Toàn bộ hệ thống chạy bằng Docker Compose

Kubernetes, tracing, retry topic và DLQ chưa bắt buộc trong MVP.

---

# 29. Version 1 Definition

Version 1 hoàn thành khi có thêm:

* Database riêng cho từng service
* Transactional Outbox
* Idempotent Consumer
* Retry mechanism
* Dead Letter Topic
* Kubernetes deployment
* Health checks
* Prometheus metrics
* Grafana dashboard
* Distributed tracing
* Graceful shutdown
* Integration test
* End-to-end test

---

# 30. Những chức năng không nằm trong phạm vi

Để tránh dự án trở nên quá lớn, phiên bản đầu không triển khai:

* Thanh toán thật
* Shopping cart
* Voucher
* Shipping provider
* Upload ảnh
* Product search
* Recommendation system
* Inventory reservation nâng cao
* Multi-warehouse
* Refund
* Web frontend
* Mobile application
* Service mesh
* Event sourcing
* CQRS đầy đủ
* Multi-region deployment

Những chức năng này chỉ được thêm sau khi toàn bộ Version 1 đã hoàn thành.

---

# 31. Tiêu chí hoàn thành dự án

Dự án được xem là hoàn chỉnh khi có thể demo rõ ràng:

* Authentication hoạt động
* Authorization hoạt động
* Redis cache hoạt động
* Kafka producer hoạt động
* Kafka consumer hoạt động
* Consumer group hoạt động
* Event ordering theo Order ID hoạt động
* Outbox tránh mất event
* Idempotency tránh duplicate notification
* Retry và DLQ hoạt động
* Docker Compose chạy toàn hệ thống
* Kubernetes deploy thành công
* Health checks hoạt động
* Logs có Request ID và Trace ID
* Prometheus thu thập được metrics
* Grafana hiển thị dashboard
* OpenTelemetry hiển thị được trace
* Hệ thống chịu được việc restart Notification Service
* Event vẫn được xử lý sau khi consumer hoạt động trở lại

---

# 32. Kết quả học tập mong đợi

Sau dự án này, người phát triển có thể giải thích:

* Monolith khác microservices như thế nào
* Tại sao mỗi service nên sở hữu database riêng
* Khi nào dùng HTTP và khi nào dùng Kafka
* Kafka topic, partition, offset và consumer group hoạt động ra sao
* Tại sao message có thể bị xử lý nhiều lần
* Tại sao cần Idempotent Consumer
* Tại sao database transaction không thể bao phủ trực tiếp Kafka
* Transactional Outbox giải quyết vấn đề gì
* Redis cache có thể gây stale data như thế nào
* Kubernetes scale stateless service ra sao
* Readiness khác liveness như thế nào
* Distributed tracing giúp debug hệ thống phân tán ra sao

---

# 33. Project summary

OrderFlow là một hệ thống đặt hàng microservices nhỏ nhưng bao phủ gần như toàn bộ các kỹ năng backend và DevOps quan trọng.

Công nghệ chính:

* Golang cho backend service
* PostgreSQL cho persistent data
* Redis cho cache, rate limiting và token management
* Kafka cho event-driven communication
* Docker cho containerization
* Kubernetes cho orchestration
* Prometheus và Grafana cho metrics
* OpenTelemetry cho distributed tracing

Luồng quan trọng nhất của hệ thống:

Người dùng tạo đơn hàng.

Order Service lưu order và outbox event trong cùng một database transaction.

Outbox Worker publish order.created lên Kafka.

Notification Service consume event.

Notification Service kiểm tra Event ID để tránh xử lý trùng.

Notification được lưu vào database.

Người dùng có thể xem notification thông qua API.

Dự án được phát triển theo từng phase, bắt đầu từ business flow đơn giản, sau đó mới bổ sung microservices, Kafka reliability, Kubernetes và observability.

Mục tiêu cuối cùng không chỉ là làm cho hệ thống chạy được, mà còn hiểu hệ thống hoạt động như thế nào khi xảy ra lỗi, retry, duplicate message, service restart và network failure.
