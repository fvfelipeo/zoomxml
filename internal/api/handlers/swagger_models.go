package handlers

import "time"

// SwaggerUser representa um usuário para documentação Swagger
type SwaggerUser struct {
	ID        int64     `json:"id" example:"1"`
	Name      string    `json:"name" example:"João Silva"`
	Email     string    `json:"email" example:"joao@exemplo.com"`
	Role      string    `json:"role" example:"user"`
	Active    bool      `json:"active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2025-08-17T19:01:44Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-08-17T19:01:44Z"`
}

// SwaggerUserWithToken representa um usuário com token para criação
type SwaggerUserWithToken struct {
	ID        int64     `json:"id" example:"1"`
	Name      string    `json:"name" example:"João Silva"`
	Email     string    `json:"email" example:"joao@exemplo.com"`
	Role      string    `json:"role" example:"user"`
	Active    bool      `json:"active" example:"true"`
	Token     string    `json:"token" example:"U6HGHy4SDK"`
	CreatedAt time.Time `json:"created_at" example:"2025-08-17T19:01:44Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-08-17T19:01:44Z"`
}

// SwaggerCompany representa uma empresa para documentação Swagger
type SwaggerCompany struct {
	ID        int64  `json:"id" example:"1"`
	Name      string `json:"name" example:"S. E. L. DE SOUZA SUARES VEICULOS"`
	CNPJ      string `json:"cnpj" example:"34.194.865/0001-58"`
	TradeName string `json:"trade_name" example:"MULTICAR VEICULOS"`

	// Endereço
	Address    string `json:"address" example:"R PERNAMBUCO"`
	Number     string `json:"number" example:"757"`
	Complement string `json:"complement" example:"LETRA A"`
	District   string `json:"district" example:"CENTRO"`
	City       string `json:"city" example:"IMPERATRIZ"`
	State      string `json:"state" example:"MA"`
	ZipCode    string `json:"zip_code" example:"65.903-320"`

	// Contato
	Phone string `json:"phone" example:"(99) 9177-7746"`
	Email string `json:"email" example:"contato@multicar.com"`

	// Dados empresariais
	CompanySize        string `json:"company_size" example:"ME"`
	MainActivity       string `json:"main_activity" example:"45.12-9-02 - Comércio sob consignação de veículos automotores"`
	SecondaryActivity  string `json:"secondary_activity" example:"45.11-1-02 - Comércio a varejo de automóveis, camionetas e utilitários usados"`
	LegalNature        string `json:"legal_nature" example:"213-5 - Empresário (Individual)"`
	OpeningDate        string `json:"opening_date" example:"12/07/2019"`
	RegistrationStatus string `json:"registration_status" example:"ATIVA"`

	// Configurações
	Restricted bool      `json:"restricted" example:"false"`
	AutoFetch  bool      `json:"auto_fetch" example:"false"`
	Active     bool      `json:"active" example:"true"`
	CreatedAt  time.Time `json:"created_at" example:"2025-08-17T19:01:44Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2025-08-17T19:01:44Z"`
}

// SwaggerError representa uma resposta de erro
type SwaggerError struct {
	Error string `json:"error" example:"Mensagem de erro"`
}

// SwaggerValidationError representa um erro de validação
type SwaggerValidationError struct {
	Error   string            `json:"error" example:"Validation failed"`
	Details map[string]string `json:"details" example:"{\"email\":\"email is required\"}"`
}

// SwaggerUsersResponse representa a resposta da listagem de usuários
type SwaggerUsersResponse struct {
	Users      []SwaggerUser     `json:"users"`
	Pagination SwaggerPagination `json:"pagination"`
}

// SwaggerCompaniesResponse representa a resposta da listagem de empresas
type SwaggerCompaniesResponse struct {
	Companies  []SwaggerCompany  `json:"companies"`
	Pagination SwaggerPagination `json:"pagination"`
}

// SwaggerPagination representa informações de paginação
type SwaggerPagination struct {
	Page  int `json:"page" example:"1"`
	Limit int `json:"limit" example:"20"`
	Total int `json:"total" example:"100"`
}

// SwaggerHealthResponse representa a resposta do health check
type SwaggerHealthResponse struct {
	Status    string `json:"status" example:"ok"`
	Timestamp int64  `json:"timestamp" example:"1755457391"`
	Version   string `json:"version" example:"1.0.0"`
}
