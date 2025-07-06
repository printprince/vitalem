package docs

// ===============================
// SWAGGER ENDPOINTS DOCUMENTATION
// ===============================

// ВАЖНОЕ РАЗЛИЧИЕ:
// - {id} - это primary key записи в таблице (patients.id, doctors.id, appointments.id)
// - {userID} - это foreign key user_id из таблицы users (связь между таблицами)

// Пример:
// GET /patients/123 - получить пациента с ID=123 из таблицы patients
// GET /users/456/patient - получить профиль пациента для пользователя с user_id=456

// ===============================
// АВТОРИЗАЦИЯ
// ===============================

// swagger:route POST /auth/login auth login
// Вход в систему
//
// Parameters:
//   - name: credentials
//     in: body
//     description: Учетные данные
//     required: true
//     schema:
//     type: object
//     properties:
//     email:
//     type: string
//     example: user@example.com
//     password:
//     type: string
//     example: password123
//
// Responses:
//
//	200: object{token=string,user=object}
//	401: ErrorResponse
func Login() {}

// swagger:route POST /auth/register auth register
// Регистрация пользователя
//
// Parameters:
//   - name: userData
//     in: body
//     required: true
//     schema:
//     type: object
//     properties:
//     email:
//     type: string
//     password:
//     type: string
//     role:
//     type: string
//     enum: [patient, doctor, admin]
//
// Responses:
//
//	201: object{user_id=string,message=string}
//	400: ErrorResponse
func Register() {}

// swagger:route GET /auth/user auth getCurrentUser
// Информация о текущем пользователе
//
// Security:
//   - BearerAuth: []
//
// Responses:
//
//	200: UserInfo
//	401: ErrorResponse
func GetCurrentUser() {}

// ===============================
// ПАЦИЕНТЫ - РАЗЛИЧИЕ ID vs userID
// ===============================

// swagger:route GET /patients patients getPatients
// Список пациентов (по primary key)
//
// Security:
//   - BearerAuth: []
//
// Responses:
//
//	200: object{patients=[]PatientResponse,total=integer}
//	401: ErrorResponse
func GetPatients() {}

// swagger:route GET /patients/{id} patients getPatientByID
// Профиль пациента по ID записи в таблице patients
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID пациента в таблице patients (primary key)
//     required: true
//     type: string
//     format: uuid
//
// Responses:
//
//	200: PatientResponse
//	404: ErrorResponse
func GetPatientByID() {}

// swagger:route GET /users/{userID}/patient patients getPatientByUserID
// Профиль пациента по User ID (для получения своего профиля)
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: userID
//     in: path
//     description: ID пользователя из таблицы users (foreign key в patients)
//     required: true
//     type: string
//     format: uuid
//
// Responses:
//
//	200: PatientResponse
//	404: ErrorResponse
func GetPatientByUserID() {}

// swagger:route PUT /users/{userID}/patient patients updatePatientProfile
// Обновление профиля пациента (второй этап регистрации)
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: userID
//     in: path
//     required: true
//     type: string
//   - name: patientData
//     in: body
//     required: true
//     schema:
//     $ref: '#/definitions/PatientCreateRequest'
//
// Responses:
//
//	200: PatientResponse
//	400: ErrorResponse
func UpdatePatientProfile() {}

// ===============================
// ВРАЧИ - РАЗЛИЧИЕ ID vs userID
// ===============================

// swagger:route GET /doctors doctors getDoctors
// Список врачей (публичный)
//
// Responses:
//
//	200: object{doctors=[]DoctorResponse,total=integer}
func GetDoctors() {}

// swagger:route GET /doctors/{id} doctors getDoctorByID
// Профиль врача по ID записи в таблице doctors
//
// Parameters:
//   - name: id
//     in: path
//     description: ID врача в таблице doctors (primary key)
//     required: true
//     type: string
//
// Responses:
//
//	200: DoctorResponse
//	404: ErrorResponse
func GetDoctorByID() {}

// swagger:route GET /users/{userID}/doctor doctors getDoctorByUserID
// Профиль врача по User ID (для получения своего профиля)
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: userID
//     in: path
//     description: ID пользователя из таблицы users (foreign key в doctors)
//     required: true
//     type: string
//
// Responses:
//
//	200: DoctorResponse
//	404: ErrorResponse
func GetDoctorByUserID() {}

// ===============================
// РАСПИСАНИЕ ВРАЧЕЙ
// ===============================

// swagger:route GET /appointments/schedules schedules getSchedules
// Расписания текущего врача
//
// Security:
//   - BearerAuth: []
//
// Responses:
//
//	200: object{schedules=[]ScheduleResponse}
//	403: ErrorResponse
func GetSchedules() {}

// swagger:route POST /appointments/schedules schedules createSchedule
// Создать расписание
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: scheduleData
//     in: body
//     required: true
//     schema:
//     $ref: '#/definitions/ScheduleCreateRequest'
//
// Responses:
//
//	201: ScheduleResponse
//	400: ErrorResponse
func CreateSchedule() {}

// swagger:route POST /appointments/schedules/{id}/generate-slots schedules generateSlots
// Генерировать слоты для расписания
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     required: true
//     type: string
//   - name: dateRange
//     in: body
//     required: true
//     schema:
//     type: object
//     properties:
//     start_date:
//     type: string
//     end_date:
//     type: string
//
// Responses:
//
//	200: object{message=string,slots_created=integer}
//	400: ErrorResponse
func GenerateSlots() {}

// ===============================
// ЗАПИСИ НА ПРИЕМ
// ===============================

// swagger:route GET /appointments appointments getAppointments
// Мои записи
//
// Security:
//   - BearerAuth: []
//
// Responses:
//
//	200: object{appointments=[]AppointmentResponse,total=integer}
//	401: ErrorResponse
func GetAppointments() {}

// swagger:route POST /appointments/{id}/book appointments bookAppointment
// Забронировать запись
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     required: true
//     type: string
//   - name: appointmentData
//     in: body
//     required: true
//     schema:
//     $ref: '#/definitions/BookAppointmentRequest'
//
// Responses:
//
//	200: AppointmentResponse
//	409: ErrorResponse
func BookAppointment() {}

// swagger:route GET /appointments/doctors/{id}/available-slots appointments getAvailableSlots
// Доступные слоты врача
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     required: true
//     type: string
//   - name: date
//     in: query
//     required: true
//     type: string
//
// Responses:
//
//	200: object{slots=[]AppointmentResponse}
//	404: ErrorResponse
func GetAvailableSlots() {}

// ===============================
// ФАЙЛЫ
// ===============================

// swagger:route GET /files files getFiles
// Мои файлы
//
// Security:
//   - BearerAuth: []
//
// Responses:
//
//	200: object{files=[]object,total=integer}
//	401: ErrorResponse
func GetFiles() {}

// swagger:route POST /files files uploadFile
// Загрузить файл
//
// Consumes:
//   - multipart/form-data
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: file
//     in: formData
//     required: true
//     type: file
//
// Responses:
//
//	201: object{file_id=string,url=string}
//	400: ErrorResponse
func UploadFile() {}

// swagger:route GET /public/{id} files getPublicFile
// Скачать публичный файл
//
// Parameters:
//   - name: id
//     in: path
//     required: true
//     type: string
//
// Responses:
//
//	200: "Файл"
//	404: ErrorResponse
func GetPublicFile() {}

// ===============================
// УВЕДОМЛЕНИЯ
// ===============================

// swagger:route GET /notifications notifications getNotifications
// Мои уведомления
//
// Security:
//   - BearerAuth: []
//
// Responses:
//
//	200: object{notifications=[]object,total=integer}
//	401: ErrorResponse
func GetNotifications() {}

// swagger:route POST /notifications notifications createNotification
// Создать уведомление
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: notificationData
//     in: body
//     required: true
//     schema:
//     type: object
//     properties:
//     recipient_id:
//     type: string
//     title:
//     type: string
//     message:
//     type: string
//
// Responses:
//
//	201: object{notification_id=string}
//	400: ErrorResponse
func CreateNotification() {}

// ===============================
// ФАЙЛЫ - ДОПОЛНИТЕЛЬНЫЕ ЭНДПОИНТЫ
// ===============================

// swagger:route GET /files/{id} files getFileInfo
// Получить мета-информацию файла
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID файла
//     required: true
//     type: string
//
// Responses:
//
//	200: object{id=string,filename=string,size=integer,content_type=string,uploaded_at=string}
//	401: ErrorResponse
//	404: ErrorResponse
func GetFileInfo() {}

// swagger:route DELETE /files/{id} files deleteFile
// Удалить файл
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID файла
//     required: true
//     type: string
//
// Responses:
//
//	204: "Файл удален"
//	401: ErrorResponse
//	403: ErrorResponse
//	404: ErrorResponse
func DeleteFile() {}

// swagger:route GET /files/{id}/download files downloadFile
// Скачать файл (авторизованный доступ)
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID файла
//     required: true
//     type: string
//
// Responses:
//
//	200: "Содержимое файла"
//	401: ErrorResponse
//	403: ErrorResponse
//	404: ErrorResponse
func DownloadFile() {}

// swagger:route GET /files/{id}/preview files previewFile
// Предпросмотр файла
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID файла
//     required: true
//     type: string
//
// Responses:
//
//	200: "Предпросмотр файла"
//	401: ErrorResponse
//	403: ErrorResponse
//	404: ErrorResponse
func PreviewFile() {}

// swagger:route PATCH /files/{id}/visibility files toggleFileVisibility
// Изменить публичность файла
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID файла
//     required: true
//     type: string
//   - name: visibilityData
//     in: body
//     required: true
//     schema:
//     type: object
//     properties:
//     is_public:
//     type: boolean
//     description: Сделать файл публичным (true) или приватным (false)
//
// Responses:
//
//	200: object{message=string,is_public=boolean}
//	401: ErrorResponse
//	403: ErrorResponse
//	404: ErrorResponse
func ToggleFileVisibility() {}

// ===============================
// ВРАЧИ - ДОПОЛНИТЕЛЬНЫЕ ЭНДПОИНТЫ
// ===============================

// swagger:route POST /doctors doctors createDoctorProfile
// Создать профиль врача (второй этап регистрации)
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: doctorData
//     in: body
//     required: true
//     schema:
//     $ref: '#/definitions/DoctorCreateRequest'
//
// Responses:
//
//	201: DoctorResponse
//	400: ErrorResponse
//	401: ErrorResponse
//	403: ErrorResponse
func CreateDoctorProfile() {}

// swagger:route PUT /doctors/{id} doctors updateDoctorByID
// Обновить профиль врача по ID записи
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID врача в таблице doctors
//     required: true
//     type: string
//   - name: doctorData
//     in: body
//     required: true
//     schema:
//     $ref: '#/definitions/DoctorCreateRequest'
//
// Responses:
//
//	200: DoctorResponse
//	400: ErrorResponse
//	401: ErrorResponse
//	403: ErrorResponse
//	404: ErrorResponse
func UpdateDoctorByID() {}

// swagger:route PUT /users/{userID}/doctor doctors updateDoctorByUserID
// Обновить профиль врача по User ID (второй этап регистрации)
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: userID
//     in: path
//     description: ID пользователя из таблицы users
//     required: true
//     type: string
//   - name: doctorData
//     in: body
//     required: true
//     schema:
//     $ref: '#/definitions/DoctorCreateRequest'
//
// Responses:
//
//	200: DoctorResponse
//	400: ErrorResponse
//	401: ErrorResponse
//	403: ErrorResponse
func UpdateDoctorByUserID() {}

// ===============================
// АВТОРИЗАЦИЯ - ДОПОЛНИТЕЛЬНЫЕ ЭНДПОИНТЫ
// ===============================

// swagger:route POST /auth/validate auth validateToken
// Валидация JWT токена
//
// Security:
//   - BearerAuth: []
//
// Responses:
//
//	200: object{valid=boolean,user_id=string,role=string}
//	401: ErrorResponse
func ValidateToken() {}

// ===============================
// ПАЦИЕНТЫ - ДОПОЛНИТЕЛЬНЫЕ ЭНДПОИНТЫ
// ===============================

// swagger:route POST /patients patients createPatientProfile
// Создать профиль пациента (второй этап регистрации)
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: patientData
//     in: body
//     required: true
//     schema:
//     $ref: '#/definitions/PatientCreateRequest'
//
// Responses:
//
//	201: PatientResponse
//	400: ErrorResponse
//	401: ErrorResponse
func CreatePatientProfile() {}

// ===============================
// РАСПИСАНИЕ - ДОПОЛНИТЕЛЬНЫЕ ЭНДПОИНТЫ
// ===============================

// swagger:route PUT /appointments/schedules/{id} schedules updateSchedule
// Обновить расписание
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID расписания
//     required: true
//     type: string
//   - name: scheduleData
//     in: body
//     required: true
//     schema:
//     $ref: '#/definitions/ScheduleCreateRequest'
//
// Responses:
//
//	200: ScheduleResponse
//	400: ErrorResponse
//	401: ErrorResponse
//	403: ErrorResponse
//	404: ErrorResponse
func UpdateSchedule() {}

// swagger:route DELETE /appointments/schedules/{id} schedules deleteSchedule
// Удалить расписание
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID расписания
//     required: true
//     type: string
//
// Responses:
//
//	204: "Расписание удалено"
//	401: ErrorResponse
//	403: ErrorResponse
//	404: ErrorResponse
func DeleteSchedule() {}

// swagger:route GET /appointments/schedules/{id}/generated-slots schedules getGeneratedSlots
// Детали сгенерированных слотов расписания
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID расписания
//     required: true
//     type: string
//   - name: date
//     in: query
//     description: Дата в формате YYYY-MM-DD
//     type: string
//
// Responses:
//
//	200: object{slots=[]object,total=integer}
//	401: ErrorResponse
//	403: ErrorResponse
//	404: ErrorResponse
func GetGeneratedSlots() {}

// swagger:route DELETE /appointments/schedules/{id}/slots schedules deleteSlots
// Удалить слоты расписания
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID расписания
//     required: true
//     type: string
//   - name: deleteData
//     in: body
//     required: true
//     schema:
//     type: object
//     properties:
//     date:
//     type: string
//     description: Дата для удаления слотов (YYYY-MM-DD)
//
// Responses:
//
//	204: "Слоты удалены"
//	401: ErrorResponse
//	403: ErrorResponse
//	404: ErrorResponse
func DeleteSlots() {}

// swagger:route PATCH /appointments/schedules/{id}/toggle schedules toggleSchedule
// Активировать/деактивировать расписание
//
// Security:
//   - BearerAuth: []
//
// Parameters:
//   - name: id
//     in: path
//     description: ID расписания
//     required: true
//     type: string
//   - name: toggleData
//     in: body
//     required: true
//     schema:
//     type: object
//     properties:
//     is_active:
//     type: boolean
//     description: Активировать (true) или деактивировать (false) расписание
//
// Responses:
//
//	200: object{message=string,is_active=boolean}
//	401: ErrorResponse
//	403: ErrorResponse
//	404: ErrorResponse
func ToggleSchedule() {}
