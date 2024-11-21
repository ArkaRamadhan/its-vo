package controllers

import (
	"net/http"

	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/kegiatan-service/models"
	"github.com/gin-gonic/gin"

)

func RequestIndex(c *gin.Context) {

	var request []models.BookingRapat
	// Tambahkan filter untuk tidak menampilkan event dengan status "pending"
	if err := initializers.DB.Where("status = ?", "pending").Find(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)

}

func AccRequest(c *gin.Context) {

	id := c.Params.ByName("id")

	var request models.BookingRapat

	if err := initializers.DB.First(&request, id); err.Error != nil {
		c.JSON(404, gin.H{"message": "Request tidak ditemukan"})
		return
	}

	// Cek apakah statusnya "pending"
	if request.Status != "pending" {
		c.JSON(400, gin.H{"message": "Request status tidak pending"})
		return
	}

	initializers.DB.Model(&request).Update("status", "acc")

	c.JSON(200, gin.H{"message": "Request diterima"})

}

func CancelRequest(c *gin.Context) {

	id := c.Params.ByName("id")

	var request models.BookingRapat

	if err := initializers.DB.First(&request, id); err.Error != nil {
		c.JSON(404, gin.H{"message": "Request tidak ditemukan"})
		return
	}

	if err := initializers.DB.Delete(&request).Error; err != nil {
		c.JSON(400, gin.H{"message": "gagal menghapus request: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Request berhasil dihapus"})

}
