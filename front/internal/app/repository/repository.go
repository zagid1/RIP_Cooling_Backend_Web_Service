// repository.go
package repository

import (
	"fmt"
	"math"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Component struct {
	ID          int
	Title       string
	TDP         int
	Vendor      string
	Description string
	Specs       []string
	Image       string
}

// Структура для заявки с результатами расчета

type Request struct {
	ID_request   int
	Description  string
	Result       string
	Count        int
	Components   []Component
	SquareArea   float64
	Height       float64
	CoolingPower float64
	Volume       float64
}

// Карта для хранения заявок (аналог taskList)
var requestList = map[int]Request{
	1: {
		ID_request: 1,
		Components: []Component{
			{ID: 1},
			{ID: 4},
		},
		Count:        2,
		SquareArea:   25.0,
		Height:       3.0,
		Volume:       75.0,
		CoolingPower: 3.6,
		Result:       "Требуемая мощность охлаждения: 3.6 кВт",
		Description:  "Заявка на систему охлаждения для серверной",
	},
}

func (r *Repository) GetComponents() ([]Component, error) {
	components := []Component{
		{
			ID:          1,
			Title:       "Intel Xeon 8490",
			TDP:         350,
			Vendor:      "Intel",
			Description: "Процессор для серверных систем высшего класса",
			Specs:       []string{"60 ядер", "120 потоков", "Частота до 3.5 GHz", "LGA4677 socket"},
			Image:       "http://127.0.0.1:9000/images/intel_zeon.webp",
		},
		{
			ID:          2,
			Title:       "AMD EPYC 9654",
			TDP:         360,
			Vendor:      "AMD",
			Description: "96 ядер, 192 потока, частота до 3.7 GHz, SP5 socket",
			Specs:       []string{"96 ядер", "192 потока", "Частота до 3.7 GHz", "SP5 socket"},
			Image:       "http://127.0.0.1:9000/images/amd_epic.webp",
		},
		{
			ID:          3,
			Title:       "Intel Xeon Gold 6418N",
			TDP:         205,
			Vendor:      "Intel",
			Description: "Сбалансированный процессор для enterprise-решений",
			Specs:       []string{"24 ядра", "48 потоков", "Частота до 3.4 GHz", "LGA4677 socket"},
			Image:       "http://127.0.0.1:9000/images/intel_xeon_gold.webp",
		},
		{
			ID:          4,
			Title:       "NVIDIA H100 80GB",
			TDP:         350,
			Vendor:      "NVIDIA",
			Description: "Флагманская GPU для AI и HPC вычислений",
			Specs:       []string{"80GB HBM3", "PCIe 5.0", "Tensor Cores", "NVLink поддержка"},
			Image:       "http://127.0.0.1:9000/images/nvidia_h100.webp",
		},
		{
			ID:          5,
			Title:       "NVIDIA A100 80GB",
			TDP:         250,
			Vendor:      "NVIDIA",
			Description: "Мощная GPU для дата-центров и AI",
			Specs:       []string{"80GB HBM2e", "PCIe 4.0", "Tensor Cores", "Multi-Instance GPU"},
			Image:       "http://127.0.0.1:9000/images/nvidia_a100.webp",
		},
		{
			ID:          6,
			Title:       "NVIDIA L40S 48GB",
			TDP:         300,
			Vendor:      "NVIDIA",
			Description: "Универсальная GPU для виртуализации и рендеринга",
			Specs:       []string{"48GB GDDR6", "PCIe 4.0", "RT Cores", "Virtualization support"},
			Image:       "http://127.0.0.1:9000/images/nvidia_l40s.webp",
		},
		{
			ID:          7,
			Title:       "Supermicro PWS-1K2IP-1R",
			TDP:         150,
			Vendor:      "Supermicro",
			Description: "Надежный блок питания для серверных стоек",
			Specs:       []string{"1200W", "80+ Platinum", "Hot-swap", "Redundant"},
			Image:       "http://127.0.0.1:9000/images/supermicro.webp",
		},
		{
			ID:          8,
			Title:       "Delta Elect. DPS-2000AB",
			TDP:         180,
			Vendor:      "Delta Electronics",
			Description: "Высокоэффективный блок питания для дата-центров",
			Specs:       []string{"2000W", "80+ Titanium", "Hot-pluggable", "N+1 redundancy"},
			Image:       "http://127.0.0.1:9000/images/delta_elect.webp",
		},
		{
			ID:          9,
			Title:       "HP 800W Flex Slot Plug",
			TDP:         80,
			Vendor:      "HP",
			Description: "Компактный блок питания для blade-систем",
			Specs:       []string{"800W", "80+ Gold", "Flex slot", "Hot plug"},
			Image:       "http://127.0.0.1:9000/images/HP_800W.webp",
		},
	}

	if len(components) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return components, nil
}

func (r *Repository) GetComponentsByTitle(title string) ([]Component, error) {
	components, err := r.GetComponents()
	if err != nil {
		return []Component{}, err
	}

	var result []Component
	for _, component := range components {
		if strings.Contains(strings.ToLower(component.Title), strings.ToLower(title)) {
			result = append(result, component)
		}
	}

	return result, nil
}

func (r *Repository) GetComponent(id int) (Component, error) {
	components, err := r.GetComponents()
	if err != nil {
		return Component{}, err
	}

	for _, component := range components {
		if component.ID == id {
			return component, nil
		}
	}
	return Component{}, fmt.Errorf("компонент не найден")
}

// Метод для получения заявки (аналог GetTask)
func (r *Repository) GetRequest(id int) (Request, error) {
	request, exists := requestList[id]
	if !exists {
		return Request{}, fmt.Errorf("заявка не найдена")
	}

	// Заменяем компоненты в заявке на полные данные из базы
	var fullComponents []Component
	for _, comp := range request.Components {
		fullComp, err := r.GetComponent(comp.ID)
		if err != nil {
			return Request{}, fmt.Errorf("ошибка получения компонента %d: %w", comp.ID, err)
		}
		fullComponents = append(fullComponents, fullComp)
	}

	request.Components = fullComponents
	return request, nil
}

// CalculateCoolingPower рассчитывает мощность охлаждения на основе объема
func (r *Repository) CalculateCoolingPower(volume float64) float64 {
	// Упрощенный расчет: объем × 0.048 кВт/м³ (эквивалент 48 Вт/м³)
	coolingPower := volume * 0.048
	// Округляем до 0.1 кВт
	return math.Round(coolingPower*10) / 10
}
