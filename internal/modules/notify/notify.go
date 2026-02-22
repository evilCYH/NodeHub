package notify

import (
	"bytes"
	"html/template"

	"github.com/bestruirui/bestsub/internal/database/op"
	notifyModel "github.com/bestruirui/bestsub/internal/models/notify"
	"github.com/bestruirui/bestsub/internal/models/setting"
	_ "github.com/bestruirui/bestsub/internal/modules/notify/channel"
	"github.com/bestruirui/bestsub/internal/modules/register"
	"github.com/bestruirui/bestsub/internal/utils/desc"
	"github.com/bestruirui/bestsub/internal/utils/log"
)

type Desc = desc.Data

func SendSystemNotify(operation uint16, title string, content any) error {
	if operation&uint16(op.GetSettingInt(setting.NOTIFY_OPERATION)) == 0 {
		return nil
	}

	nt, err := op.GetNotifyTemplateByType(notifyModel.TypeMap[operation])
	if err != nil {
		log.Errorf("failed to get notify template: %v", operation)
		return err
	}

	t, err := template.New("notify").Parse(nt.Content)
	if err != nil {
		log.Errorf("failed to parse notify template: %v", err)
		return err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, content)
	if err != nil {
		log.Errorf("failed to execute notify template: %v", err)
		return err
	}

	sysNotifyID := op.GetSettingInt(setting.NOTIFY_ID)
	notifyConfig, err := op.GetNotifyByID(uint16(sysNotifyID))
	if err != nil {
		log.Errorf("failed to get notify config: %v", sysNotifyID)
		return err
	}

	notify, err := Get(notifyConfig.Type, notifyConfig.Config)
	if err != nil {
		log.Errorf("failed to get notify: %v", err)
		return err
	}

	err = notify.Init()
	if err != nil {
		log.Errorf("failed to init notify: %v", err)
		return err
	}

	err = notify.Send(title, &buf)
	if err != nil {
		log.Errorf("failed to send notify: %v", err)
		return err
	}

	return nil
}

func Get(m string, c string) (notifyModel.Instance, error) {
	return register.Get[notifyModel.Instance]("notify", m, c)
}

func GetChannels() []string {
	return register.GetList("notify")
}

func GetInfoMap() map[string][]desc.Data {
	return register.GetInfoMap("notify")
}
