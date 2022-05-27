package artisan

import (
	"github.com/PeterYangs/superAdminCore/v2/artisan"
	"github.com/PeterYangs/superAdminCore/v2/test/artisan/demo"
)

func Artisan() []artisan.Artisan {

	return []artisan.Artisan{
		new(demo.Demo),
	}
}
