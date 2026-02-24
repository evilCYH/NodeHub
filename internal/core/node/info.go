package node

import nodeModel "github.com/evilCYH/NodeHub/internal/models/node"

func RefreshInfo() {
	refreshMutex.Lock()
	defer refreshMutex.Unlock()

	for k := range subAggBuf {
		delete(subAggBuf, k)
	}
	for k := range countryAggBuf {
		delete(countryAggBuf, k)
	}

	poolMutex.RLock()
	for _, n := range pool {
		s := subAggBuf[n.Base.SubId]
		if s == nil {
			s = &infoSums{}
			subAggBuf[n.Base.SubId] = s
		}
		s.count++
		s.sumSpeedUp += uint64(n.Info.SpeedUp.Average())
		s.sumSpeedDown += uint64(n.Info.SpeedDown.Average())
		s.sumDelay += uint64(n.Info.Delay.Average())
		s.sumRisk += uint64(n.Info.Risk)

		c := countryAggBuf[n.Info.Country]
		if c == nil {
			c = &infoSums{}
			countryAggBuf[n.Info.Country] = c
		}
		c.count++
		c.sumSpeedUp += uint64(n.Info.SpeedUp.Average())
		c.sumSpeedDown += uint64(n.Info.SpeedDown.Average())
		c.sumDelay += uint64(n.Info.Delay.Average())
		c.sumRisk += uint64(n.Info.Risk)
	}
	poolMutex.RUnlock()

	for k := range subInfoMap {
		delete(subInfoMap, k)
	}
	for k := range countryInfoMap {
		delete(countryInfoMap, k)
	}

	for subID, s := range subAggBuf {
		if s.count == 0 {
			continue
		}
		subInfoMap[subID] = nodeModel.SimpleInfo{
			Count:     s.count,
			SpeedUp:   uint32(s.sumSpeedUp / uint64(s.count)),
			SpeedDown: uint32(s.sumSpeedDown / uint64(s.count)),
			Delay:     uint16(s.sumDelay / uint64(s.count)),
			Risk:      uint8(s.sumRisk / uint64(s.count)),
		}
	}
	for country, c := range countryAggBuf {
		if c.count == 0 {
			continue
		}
		countryInfoMap[country] = nodeModel.SimpleInfo{
			Count:     c.count,
			SpeedUp:   uint32(c.sumSpeedUp / uint64(c.count)),
			SpeedDown: uint32(c.sumSpeedDown / uint64(c.count)),
			Delay:     uint16(c.sumDelay / uint64(c.count)),
			Risk:      uint8(c.sumRisk / uint64(c.count)),
		}
	}
}
