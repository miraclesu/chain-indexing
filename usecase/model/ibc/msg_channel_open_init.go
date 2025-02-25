package ibc

type MsgChannelOpenInitParams struct {
	RawMsgChannelOpenInit

	ChannelID string `json:"channelId"`
}

type RawMsgChannelOpenInit struct {
	PortID  string  `mapstructure:"port_id" json:"portId"`
	Channel Channel `mapstructure:"channel" json:"channel"`
	Signer  string  `mapstructure:"signer" json:"signer"`
}
