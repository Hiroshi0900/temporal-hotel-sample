package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"temporal-hotel-sample/internal/activities"
	"temporal-hotel-sample/internal/workflows"
)

const TaskQueue = "HOTEL_BOOKING_TASK_QUEUE"

func main() {
	// Temporalクライアントの作成
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// ワーカーの作成
	w := worker.New(c, TaskQueue, worker.Options{})

	// ワークフローとアクティビティの登録
	w.RegisterWorkflow(workflows.HotelBookingSaga)
	
	// アクティビティの登録
	w.RegisterActivity(activities.HotelRoomBookingActivity)
	w.RegisterActivity(activities.CompensateHotelRoomActivity)
	w.RegisterActivity(activities.DinnerFoodBookingActivity)
	w.RegisterActivity(activities.CompensateDinnerFoodActivity)
	w.RegisterActivity(activities.ParkingBookingActivity)
	w.RegisterActivity(activities.CompensateParkingActivity)

	log.Println("Starting hotel booking worker...")
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}

	log.Println("Worker stopped")
}
