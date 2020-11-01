package utils

import (
	"bytes"
	"time"
	"net"

	"github.com/cloudflare/goflow/v3/decoders/netflowlegacy"
	flowmessage "github.com/cloudflare/goflow/v3/pb"
	"github.com/cloudflare/goflow/v3/producer"
	"github.com/prometheus/client_golang/prometheus"
)

type StateNFLegacy struct {
	Transport Transport
	Logger    Logger
}

func (s *StateNFLegacy) DecodeFlow(msg interface{}) error {
	pkt := msg.(BaseMessage)
	buf := bytes.NewBuffer(pkt.Payload)
	key := pkt.Src.String()
	samplerAddress := pkt.Src
	if samplerAddress.To4() != nil {
		samplerAddress = samplerAddress.To4()
	}

	ts := uint64(time.Now().UTC().Unix())
	if pkt.SetTime {
		ts = uint64(pkt.RecvTime.UTC().Unix())
	}

	timeTrackStart := time.Now()
	msgDec, err := netflowlegacy.DecodeMessage(buf)

	if err != nil {
		switch err.(type) {
		case *netflowlegacy.ErrorVersion:
			NetFlowErrors.With(
				prometheus.Labels{
					"router": key,
					"error":  "error_version",
				}).
				Inc()
		}
		return err
	}

	switch msgDecConv := msgDec.(type) {
	case netflowlegacy.PacketNetFlowV5:
		NetFlowStats.With(
			prometheus.Labels{
				"router":  key,
				"version": "5",
			}).
			Inc()
		NetFlowSetStatsSum.With(
			prometheus.Labels{
				"router":  key,
				"version": "5",
				"type":    "DataFlowSet",
			}).
			Add(float64(msgDecConv.Count))
	}

	var flowMessageSet []*flowmessage.FlowMessage
	flowMessageSet, err = producer.ProcessMessageNetFlowLegacy(msgDec)

	timeTrackStop := time.Now()
	DecoderTime.With(
		prometheus.Labels{
			"name": "NetFlowV5",
		}).
		Observe(float64((timeTrackStop.Sub(timeTrackStart)).Nanoseconds()) / 1000)

	for _, fmsg := range flowMessageSet {
		fmsg.TimeReceived = ts
		fmsg.SamplerAddress = samplerAddress
	}

        _, cidr, err := net.ParseCIDR("192.168.1.0/24")
        if err != nil {
          return err
        }

	for _, flowMsg := range flowMessageSet {
            src := flowMsg.GetSrcAddr()
            dst := flowMsg.GetDstAddr()

            if(cidr.Contains(src)) {
                //m.srcMap[net.IP(src).String()] = m.srcMap[net.IP(src).String()] + flowMsg.GetBytes()
		MetricTrafficBytesByHost.With(
                    prometheus.Labels{
                        "host_ip": net.IP(src).String(),
                        "type": "src",
                    }).
                    Add(float64(flowMsg.GetBytes()))
                }
                if(cidr.Contains(dst)) {
                 //       m.dstMap[net.IP(dst).String()] = m.dstMap[net.IP(dst).String()] + flowMsg.GetBytes()
                MetricTrafficBytesByHost.With(
                    prometheus.Labels{
                        "host_ip": net.IP(dst).String(),
                        "type": "dst",
                    }).
                    Add(float64(flowMsg.GetBytes()))
                }
		//                m.log.Infof("%s>%s:%s", net.IP(src), net.IP(dst), flowMsg.GetBytes())
        }

	if s.Transport != nil {
		s.Transport.Publish(flowMessageSet)
	}

	return nil
}

func (s *StateNFLegacy) FlowRoutine(workers int, addr string, port int, reuseport bool) error {
	return UDPRoutine("NetFlowV5", s.DecodeFlow, workers, addr, port, reuseport, s.Logger)
}
