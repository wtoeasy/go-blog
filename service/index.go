/**
 * Created by GoLand.
 * User: zhu
 * Email: ylsc633@gmail.com
 * Date: 2019-05-16
 * Time: 20:17
 */
package service

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/izghua/go-blog/common"
	"github.com/izghua/go-blog/conf"
	"github.com/izghua/go-blog/entity"
	"github.com/izghua/zgh"
	"html/template"
	"strconv"
)

type IndexType string

const  (
	IndexTypeOne IndexType = "tag"
	IndexTypeTwo IndexType = "cate"
	IndexTypeThree IndexType = "default"
)


func CommonData() (h gin.H,err error) {
	h = gin.H{
		"themeJs": "/static/home/assets/js",
		"themeCss": "/static/home/assets/css",
		"themeImg": "/static/home/assets/img",
		"themeFancyboxCss": "/static/home/assets/fancybox",
		"themeFancyboxJs": "/static/home/assets/fancybox",
		"themeHLightCss": "/static/home/assets/highlightjs",
		"themeHLightJs": "/static/home/assets/highlightjs",
		"themeShareCss": "/static/home/assets/css",
		"themeShareJs": "/static/home/assets/js",
		"themeArchivesJs": "/static/home/assets/js",
		"themeArchivesCss": "/static/home/assets/css",

		"tem": "defaultList",
	}
	cates,err := CateListBySort()
	if err != nil {
		zgh.ZLog().Error("message","service.Index.CommonData","err",err.Error())
		return
	}
	var catess []common.IndexCategory
	for _,v := range cates {
		c := common.IndexCategory{
			Cates: v.Cates,
			Html: template.HTML(v.Html),
		}
		catess = append(catess,c)
	}

	tags,err := AllTags()
	if err != nil {
		zgh.ZLog().Error("message","service.Index.CommonData","err",err.Error())
		return
	}

	links,err := AllLink()
	if err != nil {
		zgh.ZLog().Error("message","service.Index.CommonData","err",err.Error())
		return
	}

	system,err := IndexSystem()
	if err != nil {
		zgh.ZLog().Error("message","service.Index.CommonData","err",err.Error())
		return
	}
	h["cates"] = catess
	h["system"] = system
	h["links"] = links
	h["tags"] = tags
	return
}


func IndexPost(page string,limit string,indexType IndexType,name string) (indexPostIndex common.IndexPostList,err error)  {
	var postKey string
	switch indexType {
	case IndexTypeOne:
		postKey = conf.TagPostIndexKey
	case IndexTypeTwo:
		postKey = conf.CatePostIndexKey
	case IndexTypeThree:
		postKey = conf.PostIndexKey
		name = "default"
	default:
		postKey = conf.PostIndexKey
	}

	field := ":name:" + name + ":page:" + page + ":limit:"+limit
	cacheRes,err := conf.CacheClient.HGet(postKey,field).Result()
	if err == redis.Nil {
		// cache key does not exist
		// set data to the cache what use the cache key
		indexPostIndex,err := doCacheIndexPostList(postKey,field,indexType,name,page,limit)
		if err != nil {
			zgh.ZLog().Error("message","service.index.IndexPost","err",err.Error())
			return indexPostIndex,err
		}
		return indexPostIndex,nil
	} else if err != nil {
		zgh.ZLog().Error("message","service.index.IndexPost","err",err.Error())
		return indexPostIndex,err
	}

	if cacheRes == "" {
		// Data is  null
		// set data to the cache what use the cache key
		indexPostIndex,err := doCacheIndexPostList(postKey,field,indexType,name,page,limit)
		if err != nil {
			zgh.ZLog().Error("message","service.index.IndexPost","err",err.Error())
			return indexPostIndex,err
		}
		return indexPostIndex,nil
	}

	err = json.Unmarshal([]byte(cacheRes),&indexPostIndex)
	if err != nil {
		zgh.ZLog().Error("message","service.index.IndexPost","err",err.Error())
		indexPostIndex,err := doCacheIndexPostList(postKey,field,indexType,name,page,limit)
		if err != nil {
			zgh.ZLog().Error("message","service.index.IndexPost","err",err.Error())
			return indexPostIndex,err
		}
		return indexPostIndex,nil
	}
	return
}


func doCacheIndexPostList(cacheKey string,field string,indexType IndexType,name string,queryPage string,queryLimit string) (res common.IndexPostList,err error) {
	limit,offset := common.Offset(queryPage,queryLimit)
	queryPageInt,err := strconv.Atoi(queryPage)
	if err != nil {
		zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
		return
	}
	var postList []*common.ConsolePostList
	var postCount int64
	switch indexType {
	case IndexTypeOne:
		tag := new(entity.ZTags)
		_,err = conf.SqlServer.Where("Name = ?",name).Get(tag)
		if err != nil {
			zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
			return
		}
		postList,err = PostTagList(tag.Id,limit,offset)
		if err != nil {
			zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
			return
		}
		postCount,err = PostTagListCount(tag.Id,limit,offset)
		if err != nil {
			zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
			return
		}
	case IndexTypeTwo:
		cate := new(entity.ZCategories)
		_,err = conf.SqlServer.Where("Name = ?",name).Get(cate)
		if err != nil {
			zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
			return
		}
		postList,err = PostCateList(cate.Id,limit,offset)
		if err != nil {
			zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
			return
		}
		postCount,err = PostCateListCount(cate.Id,limit,offset)
		if err != nil {
			zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
			return
		}
	case IndexTypeThree:
		postList,err = ConsolePostIndex(limit,offset,false)
		if err != nil {
			zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
			return
		}
		postCount,err = ConsolePostCount(limit,offset,false)
		if err != nil {
			zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
			return
		}
	default:
		postList,err = ConsolePostIndex(limit,offset,false)
		if err != nil {
			zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
			return
		}

		postCount,err = ConsolePostCount(limit,offset,false)
		if err != nil {
			zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
			return
		}
	}

	paginate := common.MyPaginate(postCount,limit,queryPageInt)

	res = common.IndexPostList{
		PostListArr: postList,
		Paginate: paginate,
	}

	jsonRes,err := json.Marshal(&res)
	if err != nil {
		zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
		return
	}
	err = conf.CacheClient.HSet(cacheKey,field,jsonRes).Err()
	if err != nil {
		zgh.ZLog().Error("message","service.index.doCacheIndexPostList","err",err.Error())
		return
	}
	return
}

func IndexPostDetail(postIdStr string) (postDetail common.IndexPostDetail,err error) {
	cacheKey := conf.PostDetailIndexKey
	field := ":post:id:" + postIdStr

	postIdInt,err := strconv.Atoi(postIdStr)
	if err != nil {
		zgh.ZLog().Error("message","service.Index.IndexPostDetail","err",err.Error())
		return
	}
	cacheRes,err := conf.CacheClient.HGet(cacheKey,field).Result()
	if err == redis.Nil {
		// cache key does not exist
		// set data to the cache what use the cache key
		postDetail,err := doCacheIndexPostDetail(cacheKey,field,postIdInt)
		if err != nil {
			zgh.ZLog().Error("message","service.index.IndexPostDetail","err",err.Error())
			return postDetail,err
		}
		return postDetail,nil
	} else if err != nil {
		zgh.ZLog().Error("message","service.index.IndexPostDetail","err",err.Error())
		return postDetail,err
	}

	if cacheRes == "" {
		// Data is  null
		// set data to the cache what use the cache key
		postDetail,err = doCacheIndexPostDetail(cacheKey,field,postIdInt)
		if err != nil {
			zgh.ZLog().Error("message","service.index.IndexPostDetail","err",err.Error())
			return postDetail,err
		}
		return postDetail,nil
	}

	err = json.Unmarshal([]byte(cacheRes),&postDetail)
	if err != nil {
		zgh.ZLog().Error("message","service.index.IndexPostDetail","err",err.Error())
		postDetail,err = doCacheIndexPostDetail(cacheKey,field,postIdInt)
		if err != nil {
			zgh.ZLog().Error("message","service.index.IndexPostDetail","err",err.Error())
			return postDetail,err
		}
		return postDetail,nil
	}
	return

}

func doCacheIndexPostDetail(cacheKey string,field string ,postIdInt int) (postDetail common.IndexPostDetail,err error) {
	postDetail,err = IndexPostDetailDao(postIdInt)
	if err != nil {
		zgh.ZLog().Error("message","service.doCacheIndexPostDetail","err",err.Error())
		return
	}
	jsonRes,err := json.Marshal(&postDetail)
	if err != nil {
		zgh.ZLog().Error("message","service.index.doCacheIndexPostDetail","err",err.Error())
		return
	}
	err = conf.CacheClient.HSet(cacheKey,field,jsonRes).Err()
	if err != nil {
		zgh.ZLog().Error("message","service.index.doCacheIndexPostDetail","err",err.Error())
		return
	}
	return
}

func PostViewAdd(postIdStr string) {
	postIdInt,err := strconv.Atoi(postIdStr)
	if err != nil {
		zgh.ZLog().Error("message","service.Index.PostViewAdd","err",err.Error())
		return
	}
	_,err = conf.SqlServer.Id(postIdInt).Incr("num").Update(entity.ZPostViews{})
	if err != nil {
		zgh.ZLog().Error("message","service.Index.PostViewAdd","err",err.Error())
		return
	}
	return
}

func PostArchives() (archivesList map[string][]*entity.ZPosts,err error) {
	cacheKey := conf.ArchivesKey
	field := ":all:"

	cacheRes,err := conf.CacheClient.HGet(cacheKey,field).Result()
	if err == redis.Nil {
		// cache key does not exist
		// set data to the cache what use the cache key
		archivesList,err := doCacheArchives(cacheKey,field)
		if err != nil {
			zgh.ZLog().Error("message","service.index.PostArchives","err",err.Error())
			return archivesList,err
		}
		return archivesList,nil
	} else if err != nil {
		zgh.ZLog().Error("message","service.index.PostArchives","err",err.Error())
		return archivesList,err
	}

	if cacheRes == "" {
		// Data is  null
		// set data to the cache what use the cache key
		archivesList,err := doCacheArchives(cacheKey,field)
		if err != nil {
			zgh.ZLog().Error("message","service.index.PostArchives","err",err.Error())
			return archivesList,err
		}
		return archivesList,nil
	}

	archivesList = make(map[string][]*entity.ZPosts)
	err = json.Unmarshal([]byte(cacheRes),&archivesList)
	if err != nil {
		zgh.ZLog().Error("message","service.index.PostArchives","err",err.Error())
		archivesList,err := doCacheArchives(cacheKey,field)
		if err != nil {
			zgh.ZLog().Error("message","service.index.PostArchives","err",err.Error())
			return archivesList,err
		}
		return archivesList,nil
	}
	return
}

func doCacheArchives(cacheKey string,field string) (archivesList map[string][]*entity.ZPosts,err error) {
	posts := make([]*entity.ZPosts,0)
	err = conf.SqlServer.Where("deleted_at IS NULL OR deleted_at = ?","0001-01-01 00:00:00").Desc("created_at").Find(&posts)
	if err != nil {
		zgh.ZLog().Error("message","service.Index.doCacheArchives","err",err.Error())
		return
	}
	archivesList = make(map[string][]*entity.ZPosts)
	for _,v := range posts {
		date := v.CreatedAt.Format("2006-01")
		archivesList[date] = append(archivesList[date],v)
	}

	jsonRes,err := json.Marshal(&archivesList)
	if err != nil {
		zgh.ZLog().Error("message","service.index.doCacheArchives","err",err.Error())
		return
	}
	err = conf.CacheClient.HSet(cacheKey,field,jsonRes).Err()
	if err != nil {
		zgh.ZLog().Error("message","service.index.doCacheArchives","err",err.Error())
		return
	}
	return
}