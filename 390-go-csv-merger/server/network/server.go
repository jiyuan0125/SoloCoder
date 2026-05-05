package network

import (
	"fmt"
	"log"
	"net"
	"sync"

	"csv-merger/common"
	"csv-merger/server/csvproc"
)

type Server struct {
	host     string
	port     string
	listener net.Listener
	wg       sync.WaitGroup
	quit     chan struct{}
}

func NewServer(host, port string) *Server {
	return &Server{
		host: host,
		port: port,
		quit: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.listener = listener
	log.Printf("Server started on %s", addr)

	go s.acceptConnections()

	return nil
}

func (s *Server) acceptConnections() {
	for {
		select {
		case <-s.quit:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.quit:
					return
				default:
					log.Printf("Failed to accept connection: %v", err)
					continue
				}
			}

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	log.Printf("New connection from %s", conn.RemoteAddr())

	var rawRequest map[string]interface{}
	if err := common.ReceiveMessage(conn, &rawRequest); err != nil {
		log.Printf("Failed to receive message: %v", err)
		return
	}

	cmd, ok := rawRequest["command"].(string)
	if !ok {
		s.sendErrorResponse(conn, "invalid command")
		return
	}

	switch common.CommandType(cmd) {
	case common.CmdPing:
		s.handlePing(conn)
	case common.CmdMergeCSV:
		s.handleMergeCSV(conn, rawRequest)
	default:
		s.sendErrorResponse(conn, fmt.Sprintf("unknown command: %s", cmd))
	}
}

func (s *Server) handlePing(conn net.Conn) {
	response := common.PingResponse{
		Success: true,
		Message: "pong",
	}

	if err := common.SendMessage(conn, response); err != nil {
		log.Printf("Failed to send ping response: %v", err)
	}
}

func (s *Server) handleMergeCSV(conn net.Conn, rawRequest map[string]interface{}) {
	inputFiles := toStringSlice(rawRequest["input_files"])
	outputFile := toString(rawRequest["output_file"])
	dedupColumns := toStringSlice(rawRequest["dedup_columns"])
	sortColumn := toString(rawRequest["sort_column"])
	sortOrder := common.SortOrder(toString(rawRequest["sort_order"]))
	customHeader := toStringSlice(rawRequest["custom_header"])

	if sortOrder == "" {
		sortOrder = common.SortAsc
	}

	request := &common.MergeRequest{
		Command:      common.CmdMergeCSV,
		InputFiles:   inputFiles,
		OutputFile:   outputFile,
		DedupColumns: dedupColumns,
		SortColumn:   sortColumn,
		SortOrder:    sortOrder,
		CustomHeader: customHeader,
	}

	response, err := csvproc.ProcessMergeRequest(request)
	if err != nil {
		log.Printf("Error processing merge request: %v", err)
		s.sendErrorResponse(conn, err.Error())
		return
	}

	if err := common.SendMessage(conn, response); err != nil {
		log.Printf("Failed to send merge response: %v", err)
	}

	if response.Success {
		log.Printf("Merge completed: %s", response.Message)
	} else {
		log.Printf("Merge failed: %s", response.Error)
	}
}

func (s *Server) sendErrorResponse(conn net.Conn, message string) {
	response := common.MergeResponse{
		Success: false,
		Error:   message,
	}
	common.SendMessage(conn, response)
}

func (s *Server) Stop() {
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
	log.Println("Server stopped")
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toStringSlice(v interface{}) []string {
	if v == nil {
		return nil
	}
	if slice, ok := v.([]interface{}); ok {
		result := make([]string, 0, len(slice))
		for _, item := range slice {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}
