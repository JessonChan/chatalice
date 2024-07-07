<script>
import { ref, onMounted, computed, nextTick } from 'vue';
// 假设 Greet 函数是从一个名为 'wails' 的模块导入的
import { Greet } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime'



export default {
  name: 'App',
  setup() {
    const chats = ref([
      {
        title: "Untitled",
        messages: [
        ]
      },
    ]);
    const currentChatIndex = ref(0);
    const userInput = ref('');
    const messageContainer = ref(null);
    const menuItems = ref([])

    const currentChat = computed(() => chats.value[currentChatIndex.value]);

    const scrollToBottom = () => {
      nextTick(() => {
        if (messageContainer.value) {
          messageContainer.value.scrollTop = messageContainer.value.scrollHeight;
        }
      });
    };

    const sendMessage = () => {
      if (userInput.value.trim() !== '') {
        currentChat.value.messages.push({ text: userInput.value.trim(), isUser: true });
        // TODO use struct directly
        Greet(JSON.stringify(currentChat.value.messages))
          .then(response => {
            currentChat.value.messages.push({ text: response, isUser: false });
            scrollToBottom();
          })
          .catch(error => {
            console.error('Error:', error);
          });
        userInput.value = '';
      }
    };

    const handleKeyDown = (event) => {
      if ((event.key === 'Enter' && event.ctrlKey && !navigator.platform.toUpperCase().includes('MAC')) ||
        (event.key === 'Enter' && event.metaKey && navigator.platform.toUpperCase().includes('MAC'))) {
        sendMessage();
      }
    }

    onMounted(() => {
      menuItems.value = [
        { icon: 'fas fa-plus', text: 'New Chat', onClickMethod: newChat },
        { icon: 'fas fa-cog', text: 'Settings', onClickMethod: newChat },
        { icon: 'fas fa-info-circle', text: 'About', onClickMethod: newChat },
      ],
        window.addEventListener('keydown', handleKeyDown);

      // Clean up the event listener on unmount
      return () => {
        window.removeEventListener('keydown', handleKeyDown);
      }
    });

    const newChat = () => {
      chats.value.unshift({ title: "Untitled", messages: [] });
      currentChatIndex.value = 0;
    };


    const selectChat = (index) => {
      currentChatIndex.value = index;
    };

    const markdownToHtml = (markdownText) => {
      // TODO 使用库将 Markdown 转换为 HTML
      return markdownText;
    };


    // 使用 window.wails.Events.On 监听事件 (需要根据实际情况调整)
    EventsOn("addMessage", (message) => {
      // 找到当前聊天的 messages 数组，并将新消息添加进去
      currentChat.value.messages.push({ text: message, isUser: false });
    });
    EventsOn("appendMessage", (message) => {
      const lastMessage = currentChat.value.messages[currentChat.value.messages.length - 1];
      if (lastMessage) {
        lastMessage.text += message;
        scrollToBottom();
      }
    });


    return {
      chats,
      currentChatIndex,
      currentChat,
      userInput,
      messageContainer,
      sendMessage,
      menuItems,
      newChat,
      selectChat,
      markdownToHtml,
    };
  }
};
</script>

<template>
  <div class="flex h-screen">
    <!-- Sidebar -->
    <div class="flex-col w-64 bg-gray-100 border-r border-gray-200">
      <div class="flex items-center justify-center h-16 border-b border-gray-200">
        <img src="./assets/images/appicon.png" alt="ChatAlice logo" class="h-6 w-6">
        <span class="text-xl font-semibold ps-2">ChatAlice</span>
      </div>
      <div class="flex-col-1 overflow-y-auto h-[calc(2/3*100vh-64px)]">
        <div class="p-4">
          <ul>
            <li v-for="(chat, index) in chats" :key="index" class="mb-2">
              <div
                :class="['flex items-center p-2 cursor-pointer rounded', currentChatIndex === index ? 'bg-gray-200' : '']"
                @click="selectChat(index)">
                <i class="fas fa-file-alt mr-2"></i>
                <span>{{ chat.title }}</span>
              </div>
            </li>
          </ul>
        </div>
      </div>
      <div class="p-4 border-t border-gray-200 h-[33%] flex flex-col justify-end">
        <div v-for="(item, index) in menuItems" :key="index" class="flex items-center mb-2 p-2 cursor-pointer"
          @click="item.onClickMethod()">
          <i :class="item.icon" class="mr-2"></i>
          <span>{{ item.text }}</span>
        </div>
      </div>
    </div>
    <!-- Main Content -->
    <div class="flex-1 flex flex-col">
      <div class="flex items-center justify-between h-16 px-4 border-b border-gray-200">
        <div class="flex items-center">
          <span class="text-lg font-medium">{{ currentChat.title }}</span>
          <span class="ml-2 px-2 py-1 text-xs bg-orange-200 text-orange-800 rounded">gpt-4.0</span>
        </div>

      </div>
      <div ref="messageContainer" class="flex-1 p-4 overflow-y-auto message-scroll">
        <div v-for="(msg, index) in currentChat.messages" :key="index" class="mb-4">
          <div class="flex items-start ">
            <div class="flex-shrink-0 mr-3">
              <i
                :class="['fas w-6', msg.isUser ? 'fa-user' : 'fa-robot', 'text-2xl', msg.isUser ? 'text-blue-500' : 'text-green-500']"></i>
            </div>
            <div class="flex-grow">
              <!-- <p class="text-gray-800 text-left">{{ markdown.render(msg.text) }}</p> -->
              <p class="text-gray-800 text-left" v-html="markdownToHtml(msg.text)"></p>
            </div>
          </div>
        </div>
      </div>
      <div class="flex items-center p-4 border-t border-gray-200" style="height: 33.33%;">
        <div class="flex-1 flex items-center relative" style="height: 100%;">
          <textarea v-model="userInput" placeholder="Type your question here..."
            class="w-full h-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none"></textarea>
          <div class="floating-icons">
            <div class="absolute inset-y-0 right-0 flex items-center space-x-4 pr-4">
              <i class="fas fa-plus-square text-gray-500 cursor-pointer"></i>
              <i class="fas fa-image text-gray-500 cursor-pointer"></i>
              <i class="fas fa-file-alt text-gray-500 cursor-pointer"></i>
              <i class="fas fa-folder-open text-gray-500 cursor-pointer"></i>
              <i class="fas fa-paperclip text-gray-500 cursor-pointer"></i>
              <i :class="['fas', 'fa-paper-plane', 'text-blue-500', 'cursor-pointer', { 'disabled': !userInput.trim() }]"
                @click="sendMessage"></i>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
/* 使用 @import 引入外部 CSS */
@import 'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.3/css/all.min.css';
@import 'https://fonts.googleapis.com/css2?family=Roboto:wght@400;500;700&display=swap';


body {
  font-family: 'Roboto', sans-serif;
}

.disabled {
  pointer-events: none;
  opacity: 0.5;
}

.floating-icons {
  position: absolute;
  bottom: 0.5rem;
  right: 0.5rem;
  display: flex;
  gap: 0.5rem;
}

.message-scroll {
  scrollbar-width: thin;
  scrollbar-color: #CBD5E0 #EDF2F7;
}

.message-scroll::-webkit-scrollbar {
  width: 8px;
}

.message-scroll::-webkit-scrollbar-track {
  background: #EDF2F7;
}

.message-scroll::-webkit-scrollbar-thumb {
  background-color: #CBD5E0;
  border-radius: 4px;
}
</style>