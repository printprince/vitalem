package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// SwaggerHandlers содержит аннотированные handler функции для Swagger документации
type SwaggerHandlers struct {
	proxyHandler *ProxyHandler
}

// NewSwaggerHandlers создает новый экземпляр SwaggerHandlers
func NewSwaggerHandlers(proxyHandler *ProxyHandler) *SwaggerHandlers {
	return &SwaggerHandlers{
		proxyHandler: proxyHandler,
	}
}

// ===============================
// АВТОРИЗАЦИЯ
// ===============================

// Login godoc
// @Summary      Вход в систему
// @Description  Аутентификация пользователя по email и паролю. Возвращает JWT токен для последующих запросов.
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Param        credentials  body      object{email=string,password=string}   true  "Данные для входа"
// @Success      200          {object}  object{token=string}  "Успешный вход"
// @Failure      400          {object}  object{error=string}  "Неверный формат данных"
// @Failure      401          {object}  object{error=string}  "Неверный email или пароль"
// @Failure      500          {object}  object{error=string}  "Внутренняя ошибка сервера"
// @Router       /auth/login [post]
func (h *SwaggerHandlers) Login(c echo.Context) error {
	return h.proxyHandler.ProxyToIdentity(c)
}

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Создание нового аккаунта пользователя (пациент или врач)
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Param        userData  body      object{email=string,password=string,role=string}  true  "Данные для регистрации"
// @Success      201       {object}  object{message=string}  "Пользователь создан"
// @Failure      400       {object}  object{error=string}    "Неверные данные или email уже существует"
// @Failure      500       {object}  object{error=string}    "Внутренняя ошибка сервера"
// @Router       /auth/register [post]
func (h *SwaggerHandlers) Register(c echo.Context) error {
	return h.proxyHandler.ProxyToIdentity(c)
}

// ValidateToken godoc
// @Summary      Валидация JWT токена
// @Description  Проверка действительности JWT токена
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string        true  "Bearer токен"
// @Success      200            {object}  object{user_id=string,email=string,role=string}  "Токен действителен"
// @Failure      401            {object}  object{error=string} "Недействительный токен"
// @Router       /auth/validate [post]
func (h *SwaggerHandlers) ValidateToken(c echo.Context) error {
	return h.proxyHandler.ProxyToIdentity(c)
}

// GetUser godoc
// @Summary      Данные текущего пользователя
// @Description  Получение информации о текущем авторизованном пользователе
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{user_id=string,email=string,role=string}  "Данные пользователя"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Router       /auth/user [get]
func (h *SwaggerHandlers) GetUser(c echo.Context) error {
	return h.proxyHandler.ProxyToIdentity(c)
}

// ===============================
// ВРАЧИ
// ===============================

// GetDoctors godoc
// @Summary      Список врачей
// @Description  Получение списка всех активных врачей (публичный доступ)
// @Tags         Врачи
// @Accept       json
// @Produce      json
// @Param        limit   query     int     false  "Количество записей"     default(10)
// @Param        offset  query     int     false  "Смещение"               default(0)
// @Param        specialization  query  string  false  "Фильтр по специализации"
// @Success      200     {object}  object{doctors=array,total=integer}    "Список врачей"
// @Failure      500     {object}  object{error=string} "Внутренняя ошибка"
// @Router       /doctors [get]
func (h *SwaggerHandlers) GetDoctors(c echo.Context) error {
	return h.proxyHandler.ProxyToSpecialist(c)
}

// GetDoctor godoc
// @Summary      Профиль врача
// @Description  Получение детальной информации о враче (публичный доступ)
// @Tags         Врачи
// @Accept       json
// @Produce      json
// @Param        id   path      string        true  "ID врача"
// @Success      200  {object}  object{id=string,first_name=string,last_name=string,specializations=array}        "Данные врача"
// @Failure      404  {object}  object{error=string} "Врач не найден"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /doctors/{id} [get]
func (h *SwaggerHandlers) GetDoctor(c echo.Context) error {
	return h.proxyHandler.ProxyToSpecialist(c)
}

// CreateDoctor godoc
// @Summary      Создать профиль врача
// @Description  Создание нового профиля врача (только для авторизованных пользователей)
// @Tags         Врачи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        doctorData  body      object{first_name=string,last_name=string,specializations=array}  true  "Данные врача"
// @Success      201         {object}  object{id=string,message=string}               "Профиль создан"
// @Failure      400         {object}  object{error=string}        "Неверные данные"
// @Failure      401         {object}  object{error=string}        "Неавторизован"
// @Failure      500         {object}  object{error=string}        "Внутренняя ошибка"
// @Router       /doctors [post]
func (h *SwaggerHandlers) CreateDoctor(c echo.Context) error {
	return h.proxyHandler.ProxyToSpecialist(c)
}

// UpdateDoctor godoc
// @Summary      Обновить профиль врача
// @Description  Обновление существующего профиля врача
// @Tags         Врачи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id          path      string               true  "ID врача"
// @Param        doctorData  body      object{first_name=string,last_name=string}  true  "Обновленные данные"
// @Success      200         {object}  object{id=string,message=string}               "Профиль обновлен"
// @Failure      400         {object}  object{error=string}        "Неверные данные"
// @Failure      401         {object}  object{error=string}        "Неавторизован"
// @Failure      403         {object}  object{error=string}        "Нет прав доступа"
// @Failure      404         {object}  object{error=string}        "Врач не найден"
// @Failure      500         {object}  object{error=string}        "Внутренняя ошибка"
// @Router       /doctors/{id} [put]
func (h *SwaggerHandlers) UpdateDoctor(c echo.Context) error {
	return h.proxyHandler.ProxyToSpecialist(c)
}

// DeleteDoctor godoc
// @Summary      Удалить врача
// @Description  Удаление профиля врача из системы
// @Tags         Врачи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID врача"
// @Success      204  "Врач удален"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      403  {object}  object{error=string} "Нет прав доступа"
// @Failure      404  {object}  object{error=string} "Врач не найден"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /doctors/{id} [delete]
func (h *SwaggerHandlers) DeleteDoctor(c echo.Context) error {
	return h.proxyHandler.ProxyToSpecialist(c)
}

// GetDoctorByUserID godoc
// @Summary      Профиль врача по User ID
// @Description  Получение профиля врача по идентификатору пользователя
// @Tags         Врачи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userID   path      string        true  "ID пользователя"
// @Success      200      {object}  object{id=string,first_name=string,last_name=string,specializations=array}       "Данные врача"
// @Failure      401      {object}  object{error=string} "Неавторизован"
// @Failure      404      {object}  object{error=string} "Врач не найден"
// @Failure      500      {object}  object{error=string} "Внутренняя ошибка"
// @Router       /users/{userID}/doctor [get]
func (h *SwaggerHandlers) GetDoctorByUserID(c echo.Context) error {
	return h.proxyHandler.ProxyToSpecialist(c)
}

// UpdateDoctorProfile godoc
// @Summary      Обновить профиль врача (User ID)
// @Description  Обновление профиля врача по идентификатору пользователя (для второго этапа регистрации)
// @Tags         Врачи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userID      path      string               true  "ID пользователя"
// @Param        doctorData  body      object{first_name=string,last_name=string,specializations=array}  true  "Данные профиля"
// @Success      200         {object}  object{id=string,message=string}               "Профиль обновлен"
// @Failure      400         {object}  object{error=string}        "Неверные данные"
// @Failure      401         {object}  object{error=string}        "Неавторизован"
// @Failure      500         {object}  object{error=string}        "Внутренняя ошибка"
// @Router       /users/{userID}/doctor [put]
func (h *SwaggerHandlers) UpdateDoctorProfile(c echo.Context) error {
	return h.proxyHandler.ProxyToSpecialist(c)
}

// ===============================
// ПАЦИЕНТЫ
// ===============================

// GetPatients godoc
// @Summary      Список пациентов
// @Description  Получение списка пациентов (только для врачей)
// @Tags         Пациенты
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query     int     false  "Количество записей"  default(10)
// @Param        offset  query     int     false  "Смещение"            default(0)
// @Success      200     {object}  object{patients=array,total=integer}   "Список пациентов"
// @Failure      401     {object}  object{error=string} "Неавторизован"
// @Failure      403     {object}  object{error=string} "Нет прав доступа"
// @Failure      500     {object}  object{error=string} "Внутренняя ошибка"
// @Router       /patients [get]
func (h *SwaggerHandlers) GetPatients(c echo.Context) error {
	return h.proxyHandler.ProxyToPatient(c)
}

// CreatePatient godoc
// @Summary      Создать профиль пациента
// @Description  Создание нового профиля пациента
// @Tags         Пациенты
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        patientData  body      object{first_name=string,last_name=string,birth_date=string}  true  "Данные пациента"
// @Success      201          {object}  object{id=string,message=string}               "Профиль создан"
// @Failure      400          {object}  object{error=string}         "Неверные данные"
// @Failure      401          {object}  object{error=string}         "Неавторизован"
// @Failure      500          {object}  object{error=string}         "Внутренняя ошибка"
// @Router       /patients [post]
func (h *SwaggerHandlers) CreatePatient(c echo.Context) error {
	return h.proxyHandler.ProxyToPatient(c)
}

// GetPatient godoc
// @Summary      Профиль пациента
// @Description  Получение детальной информации о пациенте
// @Tags         Пациенты
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string        true  "ID пациента"
// @Success      200  {object}  object{id=string,first_name=string,last_name=string}       "Данные пациента"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      403  {object}  object{error=string} "Нет прав доступа"
// @Failure      404  {object}  object{error=string} "Пациент не найден"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /patients/{id} [get]
func (h *SwaggerHandlers) GetPatient(c echo.Context) error {
	return h.proxyHandler.ProxyToPatient(c)
}

// UpdatePatient godoc
// @Summary      Обновить профиль пациента
// @Description  Обновление существующего профиля пациента
// @Tags         Пациенты
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path      string               true  "ID пациента"
// @Param        patientData  body      object{first_name=string,last_name=string,email=string}  true  "Обновленные данные"
// @Success      200          {object}  object{id=string,message=string}               "Профиль обновлен"
// @Failure      400          {object}  object{error=string}         "Неверные данные"
// @Failure      401          {object}  object{error=string}         "Неавторизован"
// @Failure      403          {object}  object{error=string}         "Нет прав доступа"
// @Failure      404          {object}  object{error=string}         "Пациент не найден"
// @Failure      500          {object}  object{error=string}         "Внутренняя ошибка"
// @Router       /patients/{id} [put]
func (h *SwaggerHandlers) UpdatePatient(c echo.Context) error {
	return h.proxyHandler.ProxyToPatient(c)
}

// DeletePatient godoc
// @Summary      Удалить пациента
// @Description  Удаление профиля пациента из системы
// @Tags         Пациенты
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID пациента"
// @Success      204  "Пациент удален"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      403  {object}  object{error=string} "Нет прав доступа"
// @Failure      404  {object}  object{error=string} "Пациент не найден"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /patients/{id} [delete]
func (h *SwaggerHandlers) DeletePatient(c echo.Context) error {
	return h.proxyHandler.ProxyToPatient(c)
}

// GetPatientByUserID godoc
// @Summary      Профиль пациента по User ID
// @Description  Получение профиля пациента по идентификатору пользователя
// @Tags         Пациенты
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userID   path      string        true  "ID пользователя"
// @Success      200      {object}  object{id=string,first_name=string,last_name=string}       "Данные пациента"
// @Failure      401      {object}  object{error=string} "Неавторизован"
// @Failure      404      {object}  object{error=string} "Пациент не найден"
// @Failure      500      {object}  object{error=string} "Внутренняя ошибка"
// @Router       /users/{userID}/patient [get]
func (h *SwaggerHandlers) GetPatientByUserID(c echo.Context) error {
	return h.proxyHandler.ProxyToPatient(c)
}

// UpdatePatientProfile godoc
// @Summary      Обновить профиль пациента (User ID)
// @Description  Обновление профиля пациента по идентификатору пользователя (для второго этапа регистрации)
// @Tags         Пациенты
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userID       path      string               true  "ID пользователя"
// @Param        patientData  body      object{first_name=string,last_name=string,birth_date=string}  true  "Данные профиля"
// @Success      200          {object}  object{id=string,message=string}               "Профиль обновлен"
// @Failure      400          {object}  object{error=string}         "Неверные данные"
// @Failure      401          {object}  object{error=string}         "Неавторизован"
// @Failure      500          {object}  object{error=string}         "Внутренняя ошибка"
// @Router       /users/{userID}/patient/profile [put]
func (h *SwaggerHandlers) UpdatePatientProfile(c echo.Context) error {
	return h.proxyHandler.ProxyToPatient(c)
}

// ===============================
// ЗАПИСИ НА ПРИЕМ
// ===============================

// GetAppointments godoc
// @Summary      Мои записи
// @Description  Получение списка записей текущего пользователя
// @Tags         Записи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query     int     false  "Количество записей"  default(10)
// @Param        offset  query     int     false  "Смещение"            default(0)
// @Success      200     {object}  object{appointments=array,total=integer}  "Список записей"
// @Failure      401     {object}  object{error=string}    "Неавторизован"
// @Failure      500     {object}  object{error=string}    "Внутренняя ошибка"
// @Router       /appointments [get]
func (h *SwaggerHandlers) GetAppointments(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// GetAvailableSlots godoc
// @Summary      Доступные слоты врача
// @Description  Получение доступных временных слотов для записи к врачу на конкретную дату
// @Tags         Записи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true   "ID врача"
// @Param        date  query     string  true   "Дата в формате YYYY-MM-DD"  example("2024-06-15")
// @Success      200   {object}  object{slots=array}  "Доступные слоты"
// @Failure      400   {object}  object{error=string} "Неверные параметры"
// @Failure      401   {object}  object{error=string} "Неавторизован"
// @Failure      404   {object}  object{error=string} "Врач не найден"
// @Failure      500   {object}  object{error=string} "Внутренняя ошибка"
// @Router       /appointments/doctors/{id}/available-slots [get]
func (h *SwaggerHandlers) GetAvailableSlots(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// BookAppointment godoc
// @Summary      Забронировать запись
// @Description  Бронирование записи к врачу (только для пациентов)
// @Tags         Записи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id              path  string  true   "ID слота для записи"
// @Param        appointmentData body  object{appointment_type=string,patient_notes=string}  true  "Данные записи"
// @Success      200             {object}  object{id=string,start_time=string,end_time=string,status=string}  "Запись забронирована"
// @Failure      400             {object}  object{error=string} "Неверные данные"
// @Failure      401             {object}  object{error=string} "Неавторизован"
// @Failure      403             {object}  object{error=string} "Только пациенты могут бронировать"
// @Failure      409             {object}  object{error=string} "Слот уже занят"
// @Failure      500             {object}  object{error=string} "Внутренняя ошибка"
// @Router       /appointments/{id}/book [post]
func (h *SwaggerHandlers) BookAppointment(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// CancelAppointment godoc
// @Summary      Отменить запись
// @Description  Отмена записи к врачу
// @Tags         Записи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID записи"
// @Success      200  {object}  object{message=string}  "Запись отменена"
// @Failure      401  {object}  object{error=string}    "Неавторизован"
// @Failure      403  {object}  object{error=string}    "Нет прав на отмену"
// @Failure      404  {object}  object{error=string}    "Запись не найдена"
// @Failure      500  {object}  object{error=string}    "Внутренняя ошибка"
// @Router       /appointments/{id}/cancel [post]
func (h *SwaggerHandlers) CancelAppointment(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// GetAppointment godoc
// @Summary      Информация о записи
// @Description  Получение детальной информации о конкретной записи
// @Tags         Записи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID записи"
// @Success      200  {object}  object{id=string,start_time=string,end_time=string,doctor_id=string,patient_id=string,status=string}  "Данные записи"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      403  {object}  object{error=string} "Нет прав доступа"
// @Failure      404  {object}  object{error=string} "Запись не найдена"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /appointments/{id} [get]
func (h *SwaggerHandlers) GetAppointment(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// CreateSchedule godoc
// @Summary      Создать расписание
// @Description  Создание нового расписания работы врача
// @Tags         Расписание
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        scheduleData  body      object{name=string,work_days=array,start_time=string,end_time=string,slot_duration=integer}  true  "Данные расписания"
// @Success      201           {object}  object{id=string,name=string,work_days=array}  "Расписание создано"
// @Failure      400           {object}  object{error=string} "Неверные данные"
// @Failure      401           {object}  object{error=string} "Неавторизован"
// @Failure      403           {object}  object{error=string} "Только врачи могут создавать расписания"
// @Failure      500           {object}  object{error=string} "Внутренняя ошибка"
// @Router       /appointments/schedules [post]
func (h *SwaggerHandlers) CreateSchedule(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// GetSchedules godoc
// @Summary      Мои расписания
// @Description  Получение всех расписаний текущего врача
// @Tags         Расписание
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{schedules=array}  "Список расписаний"
// @Failure      401  {object}  object{error=string}     "Неавторизован"
// @Failure      403  {object}  object{error=string}     "Только врачи могут просматривать расписания"
// @Failure      500  {object}  object{error=string}     "Внутренняя ошибка"
// @Router       /appointments/schedules [get]
func (h *SwaggerHandlers) GetSchedules(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// UpdateSchedule godoc
// @Summary      Обновить расписание
// @Description  Обновление существующего расписания врача
// @Tags         Расписание
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id            path  string  true   "ID расписания"
// @Param        scheduleData  body  object{name=string,work_days=array,start_time=string,end_time=string}  true  "Обновленные данные"
// @Success      200           {object}  object{id=string,name=string,updated_at=string}  "Расписание обновлено"
// @Failure      400           {object}  object{error=string} "Неверные данные"
// @Failure      401           {object}  object{error=string} "Неавторизован"
// @Failure      403           {object}  object{error=string} "Нет прав доступа"
// @Failure      404           {object}  object{error=string} "Расписание не найдено"
// @Failure      500           {object}  object{error=string} "Внутренняя ошибка"
// @Router       /appointments/schedules/{id} [put]
func (h *SwaggerHandlers) UpdateSchedule(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// DeleteSchedule godoc
// @Summary      Удалить расписание
// @Description  Удаление расписания врача
// @Tags         Расписание
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID расписания"
// @Success      200  {object}  object{message=string}  "Расписание удалено"
// @Failure      401  {object}  object{error=string}    "Неавторизован"
// @Failure      403  {object}  object{error=string}    "Нет прав доступа"
// @Failure      404  {object}  object{error=string}    "Расписание не найдено"
// @Failure      500  {object}  object{error=string}    "Внутренняя ошибка"
// @Router       /appointments/schedules/{id} [delete]
func (h *SwaggerHandlers) DeleteSchedule(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// GenerateSlots godoc
// @Summary      Генерировать слоты
// @Description  Автоматическая генерация временных слотов для расписания врача
// @Tags         Расписание
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id         path  string  true   "ID расписания"
// @Param        dateRange  body  object{start_date=string,end_date=string}  true  "Диапазон дат для генерации"
// @Success      200        {object}  object{slots_created=integer,message=string}  "Слоты сгенерированы"
// @Failure      400        {object}  object{error=string} "Неверные параметры"
// @Failure      401        {object}  object{error=string} "Неавторизован"
// @Failure      403        {object}  object{error=string} "Нет прав доступа"
// @Failure      404        {object}  object{error=string} "Расписание не найдено"
// @Failure      500        {object}  object{error=string} "Внутренняя ошибка"
// @Router       /appointments/schedules/{id}/generate-slots [post]
func (h *SwaggerHandlers) GenerateSlots(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// ToggleSchedule godoc
// @Summary      Активировать/деактивировать расписание
// @Description  Переключение активности расписания врача
// @Tags         Расписание
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path  string  true   "ID расписания"
// @Param        toggleData   body  object{is_active=boolean}  false  "Данные для переключения (опционально)"
// @Success      200          {object}  object{id=string,is_active=boolean,message=string}  "Расписание переключено"
// @Failure      401          {object}  object{error=string} "Неавторизован"
// @Failure      403          {object}  object{error=string} "Нет прав доступа"
// @Failure      404          {object}  object{error=string} "Расписание не найдено"
// @Failure      500          {object}  object{error=string} "Внутренняя ошибка"
// @Router       /appointments/schedules/{id}/toggle [patch]
func (h *SwaggerHandlers) ToggleSchedule(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// DeleteScheduleSlots godoc
// @Summary      Удалить слоты расписания
// @Description  Принудительное удаление всех сгенерированных слотов расписания
// @Tags         Расписание
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID расписания"
// @Success      200  {object}  object{message=string}  "Слоты удалены"
// @Failure      401  {object}  object{error=string}    "Неавторизован"
// @Failure      403  {object}  object{error=string}    "Нет прав доступа"
// @Failure      404  {object}  object{error=string}    "Расписание не найдено"
// @Failure      500  {object}  object{error=string}    "Внутренняя ошибка"
// @Router       /appointments/schedules/{id}/slots [delete]
func (h *SwaggerHandlers) DeleteScheduleSlots(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// GetGeneratedSlots godoc
// @Summary      Детали сгенерированных слотов
// @Description  Получение подробной информации о сгенерированных слотах расписания
// @Tags         Расписание
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id          path   string  true   "ID расписания"
// @Param        start_date  query  string  true   "Дата начала в формате YYYY-MM-DD"  example("2024-06-01")
// @Param        end_date    query  string  true   "Дата окончания в формате YYYY-MM-DD"  example("2024-06-30")
// @Success      200         {object}  object{schedule=object,period=object,slots=array,summary=object}  "Детали слотов"
// @Failure      400         {object}  object{error=string} "Неверные параметры"
// @Failure      401         {object}  object{error=string} "Неавторизован"
// @Failure      403         {object}  object{error=string} "Нет прав доступа"
// @Failure      404         {object}  object{error=string} "Расписание не найдено"
// @Failure      500         {object}  object{error=string} "Внутренняя ошибка"
// @Router       /appointments/schedules/{id}/generated-slots [get]
func (h *SwaggerHandlers) GetGeneratedSlots(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// AddException godoc
// @Summary      Создать исключение в расписании
// @Description  Добавление исключения в расписание врача (выходной день или изменение часов)
// @Tags         Исключения в расписании
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        exceptionData  body      object{date=string,type=string,custom_start_time=string,custom_end_time=string,reason=string}  true  "Данные исключения"
// @Success      201            {object}  object{id=string,date=string,type=string,reason=string}  "Исключение создано"
// @Failure      400            {object}  object{error=string} "Неверные данные"
// @Failure      401            {object}  object{error=string} "Неавторизован"
// @Failure      403            {object}  object{error=string} "Только врачи могут создавать исключения"
// @Failure      500            {object}  object{error=string} "Внутренняя ошибка"
// @Router       /appointments/exceptions [post]
func (h *SwaggerHandlers) AddException(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// GetDoctorExceptions godoc
// @Summary      Исключения врача
// @Description  Получение списка исключений в расписании врача за период
// @Tags         Исключения в расписании
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        start_date  query  string  true   "Дата начала в формате YYYY-MM-DD"  example("2024-06-01")
// @Param        end_date    query  string  true   "Дата окончания в формате YYYY-MM-DD"  example("2024-06-30")
// @Success      200         {object}  object{exceptions=array}  "Список исключений"
// @Failure      400         {object}  object{error=string}      "Неверные параметры"
// @Failure      401         {object}  object{error=string}      "Неавторизован"
// @Failure      403         {object}  object{error=string}      "Только врачи могут просматривать исключения"
// @Failure      500         {object}  object{error=string}      "Внутренняя ошибка"
// @Router       /appointments/exceptions [get]
func (h *SwaggerHandlers) GetDoctorExceptions(c echo.Context) error {
	return h.proxyHandler.ProxyToAppointment(c)
}

// ===============================
// УВЕДОМЛЕНИЯ
// ===============================

// GetMyNotifications godoc
// @Summary      Мои уведомления
// @Description  Получение списка уведомлений текущего пользователя
// @Tags         Уведомления
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{user_id=string,total_count=integer,notifications=array}  "Список уведомлений"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /notifications/my [get]
func (h *SwaggerHandlers) GetMyNotifications(c echo.Context) error {
	return h.proxyHandler.ProxyToNotification(c)
}

// GetNotification godoc
// @Summary      Информация об уведомлении
// @Description  Получение детальной информации о конкретном уведомлении
// @Tags         Уведомления
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID уведомления"
// @Success      200  {object}  object{id=integer,type=string,channel=string,message=string,status=string,created_at=string}  "Данные уведомления"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      404  {object}  object{error=string} "Уведомление не найдено"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /notifications/{id} [get]
func (h *SwaggerHandlers) GetNotification(c echo.Context) error {
	return h.proxyHandler.ProxyToNotification(c)
}

// CreateNotification godoc
// @Summary      Создать уведомление
// @Description  Создание нового уведомления (для системных уведомлений)
// @Tags         Уведомления
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        notificationData  body      object{type=string,channel=string,recipient_id=string,recipient=string,message=string}  true  "Данные уведомления"
// @Success      201               {object}  object{id=integer,type=string,message=string}  "Уведомление создано"
// @Failure      400               {object}  object{error=string} "Неверные данные"
// @Failure      401               {object}  object{error=string} "Неавторизован"
// @Failure      500               {object}  object{error=string} "Внутренняя ошибка"
// @Router       /notifications [post]
func (h *SwaggerHandlers) CreateNotification(c echo.Context) error {
	return h.proxyHandler.ProxyToNotification(c)
}

// MarkNotificationAsSent godoc
// @Summary      Отметить как отправленное
// @Description  Пометка уведомления как отправленного
// @Tags         Уведомления
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID уведомления"
// @Success      204  "Уведомление помечено как отправленное"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      404  {object}  object{error=string} "Уведомление не найдено"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /notifications/{id}/sent [put]
func (h *SwaggerHandlers) MarkNotificationAsSent(c echo.Context) error {
	return h.proxyHandler.ProxyToNotification(c)
}

// GetNotificationsByRecipient godoc
// @Summary      Уведомления по получателю
// @Description  Получение всех уведомлений для конкретного получателя (административная функция)
// @Tags         Уведомления
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        recipientId   path      string  true  "ID получателя"
// @Success      200           {object}  object{notifications=array,total_count=integer}  "Список уведомлений"
// @Failure      401           {object}  object{error=string} "Неавторизован"
// @Failure      403           {object}  object{error=string} "Нет прав доступа"
// @Failure      404           {object}  object{error=string} "Получатель не найден"
// @Failure      500           {object}  object{error=string} "Внутренняя ошибка"
// @Router       /notifications/recipient/{recipientId} [get]
func (h *SwaggerHandlers) GetNotificationsByRecipient(c echo.Context) error {
	return h.proxyHandler.ProxyToNotification(c)
}

// ===============================
// ФАЙЛЫ
// ===============================

// GetFiles godoc
// @Summary      Мои файлы
// @Description  Получение списка файлов текущего пользователя
// @Tags         Файлы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query     int     false  "Количество записей"  default(10)
// @Param        offset  query     int     false  "Смещение"            default(0)
// @Success      200     {object}  object{files=array,total=integer}      "Список файлов"
// @Failure      401     {object}  object{error=string} "Неавторизован"
// @Failure      500     {object}  object{error=string} "Внутренняя ошибка"
// @Router       /files [get]
func (h *SwaggerHandlers) GetFiles(c echo.Context) error {
	return h.proxyHandler.ProxyToFileServer(c)
}

// UploadFile godoc
// @Summary      Загрузить файл
// @Description  Загрузка нового файла в систему
// @Tags         Файлы
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file         formData  file    true   "Файл для загрузки"
// @Param        description  formData  string  false  "Описание файла"
// @Success      201          {object}  object{id=string,file_name=string,message=string}  "Файл загружен"
// @Failure      400          {object}  object{error=string}       "Неверный формат файла"
// @Failure      401          {object}  object{error=string}       "Неавторизован"
// @Failure      413          {object}  object{error=string}       "Файл слишком большой"
// @Failure      500          {object}  object{error=string}       "Внутренняя ошибка"
// @Router       /files [post]
func (h *SwaggerHandlers) UploadFile(c echo.Context) error {
	return h.proxyHandler.ProxyToFileServer(c)
}

// GetPublicFile godoc
// @Summary      Скачать публичный файл
// @Description  Скачивание публично доступного файла (без авторизации)
// @Tags         Файлы
// @Accept       json
// @Produce      application/octet-stream
// @Param        id   path      string  true  "ID файла"
// @Success      200  {file}    file    "Содержимое файла"
// @Failure      404  {object}  object{error=string} "Файл не найден или не является публичным"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /public/{id} [get]
func (h *SwaggerHandlers) GetPublicFile(c echo.Context) error {
	return h.proxyHandler.ProxyToFileServer(c)
}

// GetFile godoc
// @Summary      Информация о файле
// @Description  Получение метаинформации о файле
// @Tags         Файлы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID файла"
// @Success      200  {object}  object{id=string,name=string,size=integer,content_type=string,created_at=string}  "Метаданные файла"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      404  {object}  object{error=string} "Файл не найден"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /files/{id} [get]
func (h *SwaggerHandlers) GetFile(c echo.Context) error {
	return h.proxyHandler.ProxyToFileServer(c)
}

// DeleteFile godoc
// @Summary      Удалить файл
// @Description  Удаление файла из системы
// @Tags         Файлы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID файла"
// @Success      204  "Файл удален"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      403  {object}  object{error=string} "Нет прав доступа"
// @Failure      404  {object}  object{error=string} "Файл не найден"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /files/{id} [delete]
func (h *SwaggerHandlers) DeleteFile(c echo.Context) error {
	return h.proxyHandler.ProxyToFileServer(c)
}

// DownloadFile godoc
// @Summary      Скачать файл
// @Description  Скачивание файла (для авторизованных пользователей)
// @Tags         Файлы
// @Accept       json
// @Produce      application/octet-stream
// @Security     BearerAuth
// @Param        id   path      string  true  "ID файла"
// @Success      200  {file}    file    "Содержимое файла"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      403  {object}  object{error=string} "Нет прав доступа"
// @Failure      404  {object}  object{error=string} "Файл не найден"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /files/{id}/download [get]
func (h *SwaggerHandlers) DownloadFile(c echo.Context) error {
	return h.proxyHandler.ProxyToFileServer(c)
}

// PreviewFile godoc
// @Summary      Предпросмотр файла
// @Description  Получение предпросмотра файла (для изображений и документов)
// @Tags         Файлы
// @Accept       json
// @Produce      application/octet-stream
// @Security     BearerAuth
// @Param        id   path      string  true  "ID файла"
// @Success      200  {file}    file    "Предпросмотр файла"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      403  {object}  object{error=string} "Нет прав доступа"
// @Failure      404  {object}  object{error=string} "Файл не найден"
// @Failure      415  {object}  object{error=string} "Предпросмотр недоступен для данного типа файла"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /files/{id}/preview [get]
func (h *SwaggerHandlers) PreviewFile(c echo.Context) error {
	return h.proxyHandler.ProxyToFileServer(c)
}

// ToggleFileVisibility godoc
// @Summary      Изменить публичность файла
// @Description  Переключение статуса публичности файла (публичный/приватный)
// @Tags         Файлы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID файла"
// @Success      200  {object}  object{id=string,is_public=boolean,message=string}  "Статус изменен"
// @Failure      401  {object}  object{error=string} "Неавторизован"
// @Failure      403  {object}  object{error=string} "Нет прав доступа"
// @Failure      404  {object}  object{error=string} "Файл не найден"
// @Failure      500  {object}  object{error=string} "Внутренняя ошибка"
// @Router       /files/{id}/visibility [patch]
func (h *SwaggerHandlers) ToggleFileVisibility(c echo.Context) error {
	return h.proxyHandler.ProxyToFileServer(c)
}

// ===============================
// СИСТЕМНЫЕ ФУНКЦИИ (НЕ В SWAGGER)
// ===============================

// HealthCheck - системный эндпоинт, НЕ включен в Swagger документацию
func (h *SwaggerHandlers) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "gateway",
		"version": "1.0.0",
	})
}
