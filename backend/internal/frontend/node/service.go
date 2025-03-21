package frontendnode

type FrontendNodeService struct {
	FrontendNodeRepository *FrontendNodeRepository
}

func NewFrontendNodeService(frontendNodeRepository *FrontendNodeRepository) *FrontendNodeService {
	return &FrontendNodeService{
		FrontendNodeRepository: frontendNodeRepository,
	}
}

func (f *FrontendNodeService) GetAllReceivers() (any, error) {
	return f.FrontendNodeRepository.GetAllReceivers()
}

func (f *FrontendNodeService) GetAllProcessors() (any, error) {
	return f.FrontendNodeRepository.GetAllProcessors()
}

func (f *FrontendNodeService) GetAllExporters() (any, error) {
	return f.FrontendNodeRepository.GetAllExporters()
}
