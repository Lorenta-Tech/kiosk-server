package service

type AdminService struct {

}

func NewAdminService() *AdminService {
	return &AdminService{}
}

func (as *AdminService) FetchPrintJobs(){}