package artisan

import (
	"github.com/PeterYangs/superAdminCore/artisan"
	"github.com/PeterYangs/superAdminCore/test/artisan/demo"
)

func Artisan() []artisan.Artisan {

	return []artisan.Artisan{
		new(demo.Demo),
	}
}
