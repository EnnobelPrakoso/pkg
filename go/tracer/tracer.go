// Adapted & Modified from tdk

package tracer

import (
	"context"

	"github.com/ennobelprakoso/pkg/go/tracer/nr"
	"github.com/gin-gonic/gin"
)

// StartTransaction will create new nr transaction
func StartTransaction(name string) (t Transaction, ctx context.Context) {
	ctx = context.Background()
	return StartTransactionFromContext(ctx, name)
}

// StartTransactionFromContext to create a new nr transaction using existing context
func StartTransactionFromContext(ctx context.Context, name string) (Transaction, context.Context) {
	var t Transaction

	ctx = nr.StartTransactionWithName(ctx, name)
	t.md = nr.GetMetadataFromContext(ctx)
	t.ctx = ctx

	return t, ctx
}

func StartTransactionFromGinContext(ctx *gin.Context, name string) (Transaction, *gin.Context) {
	var t Transaction

	ctx = nr.StartGinTransactionWithName(ctx, name)
	t.md = nr.GetMetadataFromContext(ctx)
	t.ctx = ctx

	return t, ctx
}

// StartSpanFromContext create and start newrelic span from context
func StartSpanFromContext(ctx context.Context, name string) (Span, context.Context) {
	var span Span

	seg := nr.StartSegment(ctx, name)
	span.nrSegment = seg
	span.md = nr.GetMetadataFromContext(ctx)

	return span, ctx
}

// StartExternalSpanFromContext to start an external span.
// newrelic: span can be customized, either sql, redis, or http call span
func StartExternalSpanFromContext(ctx context.Context, name string, opt Options) (Span, context.Context) {
	var span Span

	switch opt.SpanType {
	case SpanTypeSQL:
		extraArgs, ok := opt.ExtraArgs.(sqlExtraArgs)
		if !ok {
			break
		}
		op := getOperationFromSQLQuery(extraArgs.Query)
		dsSegment := nr.StartPostgresSegmentWithDBName(
			ctx, extraArgs.Name, extraArgs.Host, extraArgs.Port, extraArgs.Query, name, op, opt.ExtraParam,
		)
		span.nrSegment = dsSegment
	case SpanTypeNSQPublish:
		topicName, ok := opt.ExtraArgs.(string)
		if !ok {
			break
		}
		seg := nr.StartNSQSegment(ctx, topicName)
		span.nrSegment = seg
	}

	span.ctx = ctx
	span.md = nr.GetMetadataFromContext(ctx)

	return span, ctx
}

// WithSQLSpan returns Options for SQL
func WithSQLSpan(query string, param map[string]interface{}) Options {
	return WithSQLSpanWithName("", "", "", query, param)
}

// WithSQLSpanPQInfo return an Option for tracer.StartExternalSpanFromContext
// but with Database connection information
func WithSQLSpanPQInfo(dbi DBConInfo, query string, param map[string]interface{}) Options {
	return WithSQLSpanWithName(dbi.Name, dbi.Host, dbi.Port, query, param)
}

// WithSQLSpanWithName returns Options for SQL but with the database name & host
func WithSQLSpanWithName(dbName, dbHost, dbPort, query string, param map[string]interface{}) Options {
	return Options{
		SpanType: SpanTypeSQL,
		ExtraArgs: sqlExtraArgs{
			DBConInfo: DBConInfo{
				Name: dbName,
				Host: dbHost,
				Port: dbPort,
			},
			Query: query,
		},
		ExtraParam: param,
	}
}
