package review

import (
	"sort"
	"strconv"
	"strings"

	"github.com/itsubaki/apstlib/model"
	"golang.org/x/net/context"

	"google.golang.org/appengine/capability"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func List(ctx context.Context) []int {
	kinds, err := datastore.Kinds(ctx)
	if err != nil {
		log.Warningf(ctx, err.Error())
		return []int{}
	}

	list := []int{}
	for _, k := range kinds {
		if !strings.HasPrefix(k, "Review_") {
			continue
		}
		id := strings.Split(k, "_")[1]
		idint, err := strconv.Atoi(id)
		if err != nil {
			log.Warningf(ctx, err.Error())
			continue
		}
		list = append(list, idint)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(list)))

	return list

}

func Select(ctx context.Context, kind string, limit int) model.ReviewList {
	if !capability.Enabled(ctx, "datastore_v3", "*") {
		log.Warningf(ctx, "datastore is currently unavailable.")
		return model.ReviewList{}
	}

	list := model.ReviewList{}
	q := datastore.NewQuery(kind).Order("-ID").Limit(limit)
	for t := q.Run(ctx); ; {
		var r model.Review
		_, err := t.Next(&r)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Warningf(ctx, err.Error())
			return list
		}
		list = append(list, r)
	}
	return list
}

func Insert(ctx context.Context, kind string, feed *model.ReviewFeed) {
	if !capability.Enabled(ctx, "datastore_v3", "write") {
		log.Warningf(ctx, "datastore is currently unavailable.")
		return
	}

	var key []*datastore.Key
	var src []interface{}

	for _, e := range feed.ReviewList {
		if len(e.Content) > 1500 {
			log.Warningf(ctx, "len(Content) is over 1500.")
			continue
		}

		rev := model.Review{
			ID:        e.ID,
			Title:     e.Title,
			Content:   e.Content,
			Author:    e.Author,
			Rating:    e.Rating,
			Version:   e.Version,
			Updated:   e.Updated,
			VoteSum:   e.VoteSum,
			VoteCount: e.VoteCount,
		}

		k := datastore.NewKey(ctx, kind, e.ID, 0, nil)
		key = append(key, k)
		src = append(src, &rev)
	}

	if _, err := datastore.PutMulti(ctx, key, src); err != nil {
		log.Warningf(ctx, err.Error())
	}
}
