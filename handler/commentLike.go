// package handler 负责处理 HTTP 请求：解析参数、调用 service、返回统一响应。
// 不直接操作数据库，所有业务逻辑委托给 service 层。
package handler

import (
	"blog/middleware"
	"blog/service"
	"blog/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CommentLikeHandler 评论点赞处理器
type CommentLikeHandler struct {
	service *service.CommentLikeService
}

// NewCommentLikeHandler 创建评论点赞处理器
func NewCommentLikeHandler() *CommentLikeHandler {
	return &CommentLikeHandler{service: service.NewCommentLikeService()}
}

// Toggle 切换评论点赞（需要登录）
// @Summary 切换评论点赞状态
// @Tags 评论点赞
// @Security BearerAuth
// @Param id path int true "评论 ID"
// @Success 200 {object} utils.Response{data=object{liked=bool,count=int}}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /comments/{id}/like [post]
func (h *CommentLikeHandler) Toggle(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "评论 ID 格式错误")
		return
	}

	liked, err := h.service.Toggle(uint(id), userID)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	count := h.service.GetLikeCount(uint(id))
	utils.Success(c, gin.H{
		"liked": liked,
		"count": count,
	})
}

// Status 获取当前用户对某条评论的点赞状态（可选登录）
// @Summary 获取评论点赞状态
// @Tags 评论点赞
// @Param id path int true "评论 ID"
// @Success 200 {object} utils.Response{data=object{liked=bool,count=int}}
// @Failure 400 {object} utils.Response
// @Router /comments/{id}/like [get]
func (h *CommentLikeHandler) Status(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "评论 ID 格式错误")
		return
	}

	count := h.service.GetLikeCount(uint(id))
	resp := gin.H{
		"count": count,
		"liked": false,
	}

	if userID, ok := middleware.GetCurrentUserID(c); ok {
		resp["liked"] = h.service.IsLiked(uint(id), userID)
	}

	utils.Success(c, resp)
}

// BatchStatus 批量获取评论点赞状态（可选登录）。
// 查询参数：ids=1,2,3
// 返回：map[commentID]{count, liked}
// @Summary 批量获取评论点赞状态
// @Tags 评论点赞
// @Param ids query string true "评论 ID，逗号分隔，例如 1,2,3"
// @Success 200 {object} utils.Response{data=map[uint]object{liked=bool,count=int}}
// @Router /comments/likes [get]
func (h *CommentLikeHandler) BatchStatus(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		utils.Success(c, gin.H{})
		return
	}

	var commentIDs []uint
	for _, s := range strings.Split(idsStr, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			continue
		}
		commentIDs = append(commentIDs, uint(id))
	}

	userID, hasUser := middleware.GetCurrentUserID(c)
	counts := h.service.BatchGetLikeCounts(commentIDs)
	var likedMap map[uint]bool
	if hasUser {
		likedMap = h.service.BatchIsLiked(commentIDs, userID)
	}

	resp := make(map[uint]gin.H, len(commentIDs))
	for _, id := range commentIDs {
		liked := false
		if hasUser {
			liked = likedMap[id]
		}
		resp[id] = gin.H{
			"count": counts[id],
			"liked": liked,
		}
	}

	utils.Success(c, resp)
}
