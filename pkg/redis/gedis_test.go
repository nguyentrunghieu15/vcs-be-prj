package redis

import (
	"context"
	"reflect"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestGedis_Set(t *testing.T) {
	type fields struct {
		clientredis *redis.Client
	}
	type args struct {
		key   string
		value map[string]string
	}

	want := redis.NewIntCmd(context.Background())
	want.SetVal(10)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *redis.IntCmd
	}{
		// TODO: Add test cases.
		{
			name: "Test set a server",
			fields: fields(*NewRedisClient(
				GedisConfig{
					Addess:   "localhost:6379",
					Password: "",
					Username: ""})),
			args: args{
				key: "list:server:2",
				value: map[string]string{
					"id":         "2",
					"created_at": "now",
					"updated_at": "now",
					"Deleted_at": "now",
					"created_by": "now",
					"updated_by": "now",
					"deleted_by": "now",
					"name":       "test",
					"ipv4":       "0.0.0.0",
					"status":     "on",
				},
			},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gedis{
				clientredis: tt.fields.clientredis,
			}
			if got := g.Set(tt.args.key, tt.args.value); !reflect.DeepEqual(got.Val(), tt.want.Val()) {
				t.Errorf("Gedis.Set() = %v, want %v", got, tt.want)
			}
			g.DelKeyWithPattern(tt.args.key)
		})
	}
}

func TestGedis_Get(t *testing.T) {
	type fields struct {
		clientredis *redis.Client
	}
	type args struct {
		key string
	}
	want := redis.NewMapStringStringCmd(context.Background())
	want.SetVal(
		map[string]string{
			"id":         "2",
			"created_at": "now",
			"updated_at": "now",
			"Deleted_at": "now",
			"created_by": "now",
			"updated_by": "now",
			"deleted_by": "now",
			"name":       "test",
			"ipv4":       "0.0.0.0",
			"status":     "on",
		},
	)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *redis.MapStringStringCmd
	}{
		// TODO: Add test cases.
		struct {
			name   string
			fields fields
			args   args
			want   *redis.MapStringStringCmd
		}{
			name: "Get a server",
			fields: fields(*NewRedisClient(
				GedisConfig{
					Addess:   "localhost:6379",
					Password: "",
					Username: ""})),
			args: args{
				key: "list:server:1",
			},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gedis{
				clientredis: tt.fields.clientredis,
			}
			g.Set(tt.args.key, tt.want.Val())
			if got := g.Get(tt.args.key); !reflect.DeepEqual(got.Val(), tt.want.Val()) {
				t.Errorf("Gedis.Get() = %v, want %v", got, tt.want)
			}
			g.DelKeyWithPattern(tt.args.key)
		})
	}
}
