package types

import (
	"reflect"
	"testing"
)

func TestParseMoveStructTag(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name    string
		args    args
		want    StructTag
		wantErr bool
	}{
		{
			name: "simple",
			args: args{"0x1::lp::LP"},
			want: StructTag{
				Address:    "0x1",
				Module:     "lp",
				Name:       "LP",
				TypeParams: []TypeTag{},
			},
			wantErr: false,
		},
		{
			name: "simple 01",
			args: args{"0x1::lp::LP<0x1::coin::Coin, 0x2::coin::BTC>"},
			want: StructTag{
				Address: "0x1",
				Module:  "lp",
				Name:    "LP",
				TypeParams: []TypeTag{
					{
						StructTag: &StructTag{
							Address:    "0x1",
							Module:     "coin",
							Name:       "Coin",
							TypeParams: []TypeTag{},
						},
					},
					{
						StructTag: &StructTag{
							Address:    "0x2",
							Module:     "coin",
							Name:       "BTC",
							TypeParams: []TypeTag{},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "simple 02",
			args: args{"0x1::lp::LP<0x1::coin::Coin, 0x1::lp::LP<0x1::coin::Coin, 0x2::coin::BTC>>"},
			want: StructTag{
				Address: "0x1",
				Module:  "lp",
				Name:    "LP",
				TypeParams: []TypeTag{
					{
						StructTag: &StructTag{
							Address:    "0x1",
							Module:     "coin",
							Name:       "Coin",
							TypeParams: []TypeTag{},
						},
					},
					{
						StructTag: &StructTag{
							Address: "0x1",
							Module:  "lp",
							Name:    "LP",
							TypeParams: []TypeTag{
								{
									StructTag: &StructTag{
										Address:    "0x1",
										Module:     "coin",
										Name:       "Coin",
										TypeParams: []TypeTag{},
									},
								},
								{
									StructTag: &StructTag{
										Address:    "0x2",
										Module:     "coin",
										Name:       "BTC",
										TypeParams: []TypeTag{},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMoveStructTag(tt.args.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMoveStructTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMoveStructTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructTag_GetFullName(t *testing.T) {
	type fields struct {
		Address    string
		Module     string
		Name       string
		TypeParams []TypeTag
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test",
			fields: fields{
				Address: "0x1",
				Module:  "lp",
				Name:    "LP",
				TypeParams: []TypeTag{
					{
						StructTag: &StructTag{
							Address:    "0x1",
							Module:     "coin",
							Name:       "Coin",
							TypeParams: []TypeTag{},
						},
					},
					{
						StructTag: &StructTag{
							Address: "0x1",
							Module:  "lp",
							Name:    "LP",
							TypeParams: []TypeTag{
								{
									StructTag: &StructTag{
										Address:    "0x1",
										Module:     "coin",
										Name:       "Coin",
										TypeParams: []TypeTag{},
									},
								},
								{
									StructTag: &StructTag{
										Address:    "0x1",
										Module:     "coin",
										Name:       "BTC",
										TypeParams: []TypeTag{},
									},
								},
							},
						},
					},
				},
			},
			want: "0x1::lp::LP<0x1::coin::Coin, 0x1::lp::LP<0x1::coin::Coin, 0x1::coin::BTC>>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &StructTag{
				Address:    tt.fields.Address,
				Module:     tt.fields.Module,
				Name:       tt.fields.Name,
				TypeParams: tt.fields.TypeParams,
			}
			if got := tr.GetFullName(); got != tt.want {
				t.Errorf("StructTag.GetFullName() = %v, want %v", got, tt.want)
			}
		})
	}
}
