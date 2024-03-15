package main

import (
	"testing"
)

func TestExtractPriceAndDescription(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    MessageData
		wantErr bool
	}{
		{
			name:    "Price and description",
			message: "20.5 for groceries",
			want:    MessageData{Price: 20.5, Description: "for groceries"},
			wantErr: false,
		},
		{
			name:    "Price only",
			message: "20.5",
			want:    MessageData{Price: 20.5, Description: ""},
			wantErr: false,
		},
		{
			name:    "No price",
			message: "groceries",
			want:    MessageData{Price: 0, Description: "groceries"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractPriceAndDescription(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractPriceAndDescription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractPriceAndDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	md := MessageData{Price: 20.5, Description: "for groceries"}
	want := "Added 20.50 (for groceries)"
	if got := md.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestDescriptionWithFallback(t *testing.T) {
	md := MessageData{Price: 20.5, Description: ""}
	want := "?"
	if got := md.DescriptionWithFallback(); got != want {
		t.Errorf("DescriptionWithFallback() = %v, want %v", got, want)
	}
}

func TestGetPriceFromString(t *testing.T) {
	tests := []struct {
		s    string
		want float64
	}{
		{
			s:    "20.5 for groceries",
			want: 20.5,
		},
		{
			s:    "groceries",
			want: -1,
		},
		{s: "R$540.00", want: 540},
		{s: "R$41.00", want: 41},
		{s: "R$73.70", want: 73.70},
		{s: "R$16.95", want: 16.95},
		{s: "R$1,090.00", want: 1090},
		{s: "R$4,359.23", want: 4359.23},
		{s: "R$100.00", want: 100},
		{s: "R$149.47		", want: 149.47},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := GetPriceFromString(tt.s); got != tt.want {
				t.Errorf("GetPriceFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
