<script setup>
import { ref, onMounted, computed, nextTick } from 'vue';
import { Call } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime';
import { WindowMaximise, WindowUnfullscreen, WindowToggleMaximise } from '../wailsjs/runtime';

import { Marked } from 'marked';
import { markedHighlight } from "marked-highlight"
import hljs from 'highlight.js'
//引入markdown样式
import 'highlight.js/styles/atom-one-dark.css'

const chats = ref([]);
const currentChatIndex = ref(0);
const userInput = ref('');
const uploadedImages = ref([])
const fullImageSrc = ref('')
const messageContainer = ref(null);
const chatContainer = ref(null);
const menuItems = ref([]);
const showSettings = ref(false);
const showAbout = ref(false);
const showSettingsList = ref(false);
const settings = ref({ name: '', key: '', baseUrl: '' });
const submittedSettings = ref([]);
const showChatSetting = ref(false);
const selectTitle= ref('');
const selectedModel = ref(null);
const conversationRounds = ref(3);
const maxInputTokens = ref(4096);
const maxOutputTokens = ref(4096);
const systemPrompt = ref('You are a helpful assistant.');
const shouldScroll = ref(true);
const messagesToShow = ref(10); // 默认显示的消息数量
const isMaximized = ref(false);
const isUploading = ref(false);
const imageError = ref('');
const MAX_IMAGE_SIZE = 5 * 1024 * 1024; // 5MB
const ALLOWED_IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];

const toggleMaximize = async () => {
  WindowToggleMaximise()
  return;
};

const submitSettings = () => {
  submittedSettings.value.push({ ...settings.value });
  Call("insertModel", JSON.stringify(settings.value));
  settings.value = { name: '', key: '', baseUrl: '', model: '' };
};

const submitChatSettings = () => {
  chats.value[currentChatIndex.value] = {
    title: selectTitle.value,
    messages: currentChat.value.messages,
    id: currentChat.value.id,
    modelId: selectedModel.value.id,
    conversationRounds: conversationRounds.value,
    maxInputTokens: maxInputTokens.value,
    maxOutputTokens: maxOutputTokens.value,
    systemPrompt: systemPrompt.value,
  }
  Call("updateChatSetting", JSON.stringify(
    {
      ...currentChat.value,
      ...{
        chatId: currentChat.value.id, messages: []
      }
    }
  ));

  showChatSetting.value = false;
  selectedModel.value = null;
  conversationRounds.value = 3
  maxInputTokens.value = 4096
  maxOutputTokens.value = 4096
  systemPrompt.value = 'You are a helpful assistant.'
};

const toggleSettingsList = () => {
  if (submittedSettings.value.length > 0) {
    // showSettingsList.value = true;
    showChatSetting.value = true;
    selectedModel.value = submittedSettings.value.find(item => item.id === currentModelId.value) ?? submittedSettings.value[0]
    selectTitle.value=currentChat.value.title
    conversationRounds.value = currentChat.value.conversationRounds ?? 3
    maxInputTokens.value = currentChat.value.maxInputTokens ?? 4096
    maxOutputTokens.value = currentChat.value.maxOutputTokens ?? 4096
    systemPrompt.value = currentChat.value.systemPrompt ?? 'You are a helpful assistant.'
  } else {
    // 先增加一个新的模型
    showSettings.value = true;
  }
};

const toggleSidebar = () => {
  document.getElementById('sidebar')?.classList.toggle('hidden');
  document.getElementById('miniSidebar')?.classList.toggle('hidden');
};

const deleteSetting = (index) => {
  submittedSettings.value.splice(index, 1);
}

const currentChat = computed(() => chats.value[currentChatIndex.value]);
const currentChatModelName = computed(() => {
  const setting = submittedSettings.value?.find(item => item.id == currentChat.value?.modelId);
  return setting ? setting.name : '';
});
const currentModelId = computed(() => {
  return currentChat.value?.modelId ?? submittedSettings.value[0]?.id;
});

const scrollToBottom = () => {
  nextTick(() => {
    console.log("scroll to bottom", shouldScroll.value)
    if (messageContainer.value && shouldScroll.value) {
      messageContainer.value.scrollTop = messageContainer.value.scrollHeight;
    }
  });
};
const stopScrolling = () => {
  shouldScroll.value = false;
};
const sendMessage = () => {
  if (userInput.value.trim() !== '' || uploadedImages.value.length > 0) {
    shouldScroll.value = true;
    currentChat.value.messages = currentChat.value.messages.slice(-8);
    
    // 准备图片数据
    const imageUrls = uploadedImages.value.map(img => img.src);
    
    currentChat.value.messages.push({ 
      text: userInput.value.trim(),
      images: [...uploadedImages.value], // 保存图片信息到消息中
      isUser: true 
    });

    Call("sendMessage", JSON.stringify({
      Content: userInput.value.trim(),
      Images: imageUrls.join("&"),
      ChatID: chats.value[currentChatIndex.value].id,
      ModelID: currentModelId.value,
    })).then(response => {
      response = JSON.parse(response);
      currentChat.value.messages.push({ 
        text: response.text, 
        isUser: false, 
        id: response.message_id 
      });
      scrollToBottom();
    }).catch(error => {
      console.error('Error:', error);
      // 添加错误处理提示
      currentChat.value.messages.push({ 
        text: "Failed to send message. Please try again.",
        isUser: false,
        isError: true
      });
    });

    userInput.value = '';
    uploadedImages.value = [];
    
    if (currentChatIndex.value > 0) {
      chats.value.unshift(chats.value.splice(currentChatIndex.value, 1)[0]);
      currentChatIndex.value = 0;
    }
  }
};

const handleImageUpload = async (event) => {
  const files = event.target.files;
  isUploading.value = true;
  imageError.value = '';
  
  try {
    for (let file of files) {
      if (file.size > MAX_IMAGE_SIZE) {
        imageError.value = `File ${file.name} exceeds 5MB limit`;
        continue;
      }
      
      if (!ALLOWED_IMAGE_TYPES.includes(file.type)) {
        imageError.value = `File ${file.name} must be JPEG, PNG, GIF or WebP`;
        continue;
      }

      const reader = new FileReader();
      await new Promise((resolve, reject) => {
        reader.onload = (e) => {
          uploadedImages.value.push({
            id: Date.now(),
            src: e.target.result,
            name: file.name,
            size: file.size
          });
          resolve();
        };
        reader.onerror = reject;
        reader.readAsDataURL(file);
      });
    }
  } catch (error) {
    imageError.value = 'Failed to upload image';
    console.error('Image upload error:', error);
  } finally {
    isUploading.value = false;
    // 清空 input 以允许重复上传相同文件
    event.target.value = '';
  }
};

const removeImage = (id) => {
  const index = uploadedImages.value.findIndex(img => img.id === id);
  if (index !== -1) {
    uploadedImages.value.splice(index, 1);
  }
};

const showFullImage = (imageSrc) => {
  fullImageSrc.value = imageSrc;
}

const handleKeyDown = (event) => {
  if ((event.key === 'Enter' && event.ctrlKey && !navigator.platform.toUpperCase().includes('MAC')) ||
    (event.key === 'Enter' && event.metaKey && navigator.platform.toUpperCase().includes('MAC'))) {
    sendMessage();
  }
};

const refreshModelList = () => {
  Call("getModelList", "").then(response => {
    let modelList = JSON.parse(response);
    if (modelList.length > 0) {
      submittedSettings.value = modelList.map(item => ({ name: item.name, key: item.key, baseUrl: item.baseUrl, model: item.model, id: item.ID }));
    }
    console.log(modelList, submittedSettings.value)
  })
}

const getChats = (isIntialLoad = false) => {
  let lastSeen = new Date().getTime() / 1000;
  if (chats.value.length > 0) {
    chats.value.forEach(item => {
      console.log(item.updatedAt, new Date(item.updatedAt * 1000), item.updatedAt < lastSeen)
      if (item.updatedAt < lastSeen) {
        lastSeen = new Date(item.updatedAt * 1000).getTime() / 1000;
      }
    })
  }
  Call("getChats", `${Math.floor(lastSeen)}`).then(data => {
    let response = JSON.parse(data);
    console.log(response);
    chats.value = [
      ...chats.value,
      ...response.map(item => ({ ...item }))
    ]
    if (chats.value.length == 0) {
      newChat();
    }
    if (isIntialLoad && chats.value.length > 0) {
      selectChat(0);
    }
  })
}
const handleChatsScroll = (event) => {
  const { scrollTop, clientHeight, scrollHeight } = event.target;
  if (scrollTop + clientHeight >= scrollHeight - 50) {
    getChats()
  }
};

onMounted(() => {
  menuItems.value = [
    { icon: 'fas fa-plus', text: 'New Chat', onClickMethod: newChat },
    {
      icon: 'fas fa-cog', text: 'Settings', onClickMethod: () => {
        showSettings.value = true;
      }
    },
    {
      icon: 'fas fa-info-circle', text: 'About', onClickMethod: () => {
        showAbout.value = true;
      }
    },
  ];
  refreshModelList();
  getChats(true);
  window.addEventListener('keydown', handleKeyDown);

  // 监听鼠标事件，停止滚动
  ['wheel', 'click'].forEach(eventName => {
    document.addEventListener(eventName, () => {
      stopScrolling();
    }, { passive: true });
  });

  // Open all links externally
  // This issue https://github.com/wailsapp/wails/issues/2691
  document.body.addEventListener('click', function (e) {
    if (e.target && e.target.nodeName == 'A' && e.target.href) {
      const url = e.target.href;
      if (
        !url.startsWith('http://#') &&
        !url.startsWith('file://') &&
        !url.startsWith('http://wails.localhost:')
      ) {
        e.preventDefault();
        window.runtime.BrowserOpenURL(url);
      }
    }
  });

  // Clean up the event listener on unmount
  return () => {
    window.removeEventListener('keydown', handleKeyDown);
  };
});

const newChat = () => {
  let newChatItem = {
    ...currentChat.value,
    ...{
      title: "Untitled",
      messages: [],
      id: new Date().getTime(),
      conversationRounds: 3,
      maxInputTokens: 4096,
      maxOutputTokens: 4096,
      systemPrompt: 'You are a helpful assistant.'
    }
  }
  chats.value.unshift(newChatItem);
  currentChatIndex.value = 0;
};

const selectChat = (index) => {
  currentChatIndex.value = index;
  shouldScroll.value = true;
  scrollToBottom();
};

const marked = new Marked(
  markedHighlight({
    langPrefix: 'hljs language-',
    highlight(code, lang) {
      const language = hljs.getLanguage(lang) ? lang : 'shell'
      return hljs.highlight(code, { language }).value
    }
  })
)

const markdownToHtml = (markdownText) => {
  // TODO 使用库将 Markdown 转换为 HTML
  return marked.parse(markdownText);
};

// 使用 window.wails.Events.On 监听事件 (需要根据实际情况调整)
EventsOn("addMessage", (message) => {
  currentChat.value.messages.push({ text: message, isUser: false });
});

EventsOn("updateChatTitle", (data) => {
  let chat = JSON.parse(data);
  chats.value.find(({ id }) => id === chat.id).title = chat.title;
});


EventsOn("appendMessage", (data) => {
  let message = JSON.parse(data);
  // console.log("message", message, currentChat.value.messages);
  // loop currentChat.messages to find the id===message.message_id and upate the text+=text
  // 从后向前查找消息
  for (let i = currentChat.value.messages.length - 1; i >= 0; i--) {
    if (currentChat.value.messages[i].id === message.message_id) {
      currentChat.value.messages[i].text += message.text;
      scrollToBottom();
      break; // 找到后退出循环
    }
  }
});
EventsOn("updateMessage", (data) => {
  console.log(data)
  let message = JSON.parse(data);
  // 从后向前查找消息
  for (let i = currentChat.value.messages.length - 1; i >= 0; i--) {
    if (currentChat.value.messages[i].id === message.message_id) {
      currentChat.value.messages[i].text = message.text;
      scrollToBottom();
      break; // 找到后退出循环
    }
  }
});

const displayedMessages = computed(() => {
  return currentChat.value?.messages.slice(-messagesToShow.value);
});

const handleScroll = (event) => {
  const { scrollTop, clientHeight, scrollHeight } = event.target;
  console.log(scrollTop, clientHeight, scrollHeight,messagesToShow.value)
  if (scrollTop === 0 && clientHeight < scrollHeight) {
    // 当滚动到顶部时，加载更多消息
    messagesToShow.value = Math.floor(messagesToShow.value * 1.5); // 确保乘积是整数
    // 强制更新视图，设置 scrollTop 为一个小的正值
    nextTick(() => {
      event.target.scrollTop = 1; // 或者设置为其他小值
    });
  } else if (scrollHeight - scrollTop === clientHeight ) {
    // 当滚动到最底部时，重置显示的消息数量
    messagesToShow.value = 10;
  }
};
</script>
<template>
  <div style="height: 12px; width: 100%;">
    <div class="flex-col w-64 bg-gray-100 border-r border-gray-200" style="min-height: 12px;">
    </div>
  </div>
  <div class="flex h-screen">
    <div id="miniSidebar" class="flex-col w-64 bg-gray-100 border-r border-gray-200 hidden">
      <div id="miniSidebar-header" class="flex w-8 items-center justify-center h-16 border-r border-gray-200"
        @dblclick="toggleMaximize" style="user-select: none;--wails-draggable:drag">
        <img src="./assets/images/appicon.png" alt="ChatAlice logo" class="h-6 w-6" style="--wails-draggable:no-drag">
      </div>
      <div class="flex items-center justify-between w-full">
        <i class="fas fa-chevron-circle-right cursor-pointer mx-auto" @click="toggleSidebar"></i>
      </div>
      <div class="flex items-center justify-between w-full py-4">
        <i class="fas fa-plus cursor-pointer mx-auto" @click="newChat"></i>
      </div>
    </div>
    <!-- Sidebar -->
    <div id="sidebar" class="flex-col w-64 bg-gray-100 border-r border-gray-200">
      <div id="sidebar-header" class="flex items-center justify-center h-16 border-r border-gray-200"
        @dblclick="toggleMaximize" style="user-select: none;--wails-draggable:drag">
        <img src="./assets/images/appicon.png" alt="ChatAlice logo" class="h-6 w-6" style="--wails-draggable:no-drag">
        <span class="text-xl font-semibold ps-2" style="--wails-draggable:no-drag">ChatAlice</span>
      </div>
      <div class="flex items-center justify-between w-full px-8">
        <i class="fas fa-bars cursor-pointer" @click="toggleSidebar"></i>
        <i class="fas fa-plus cursor-pointer" @click="newChat"></i>
      </div>

      <div ref="chatContainer" class="flex-col-1 w-64 overflow-y-auto h-[calc(2/3*100vh-64px)]"
        @scroll="handleChatsScroll">
        <div class="p-4">
          <ul>
            <li v-for="(chat, index) in chats" :key="index" class="mb-2">
              <div
                :class="['flex items-center p-2 cursor-pointer rounded', currentChatIndex === index ? 'bg-gray-200' : '']"
                @click="selectChat(index)">
                <!-- <i class="fas fa-file-alt mr-2"></i> -->
                <span class="chat-title" :title="chat.title">{{ chat.title }}</span>
              </div>
            </li>
          </ul>
        </div>
      </div>
      <div class="p-4 h-[33%] flex flex-col justify-end">
        <div v-for="(item, index) in menuItems" :key="index" class="flex items-center mb-2 p-2 cursor-pointer"
          @click="item.onClickMethod()">
          <i :class="item.icon" class="mr-2"></i>
          <span>{{ item.text }}</span>
        </div>
      </div>
    </div>
    <!-- Main Content -->
    <div id="main-content" class="flex-1 flex flex-col">
      <div id="main-content-header" class="flex items-center justify-between h-16 px-4 border-r border-gray-200"
        @dblclick="toggleMaximize" style="--wails-draggable:drag">
        <div class="flex items-center" style="--wails-draggable:no-drag">
          <span class="text-lg font-medium">{{ currentChat?.title }}</span>
          <div class="relative pl-2">
            <span @click="toggleSettingsList"
              class="bg-orange-200 text-orange-800 px-2 py-1 rounded text-sm cursor-pointer">{{ currentChatModelName
              }}</span>
          </div>
        </div>

      </div>
      <div ref="messageContainer" class="flex-1 p-4 overflow-y-auto message-scroll" @scroll="handleScroll">
        <div v-for="(msg, index) in displayedMessages" :key="index" class="mb-4">
          <div class="flex items-start ">
            <div class="flex-shrink-0 mr-3">
              <i
                :class="['fas w-6', msg.isUser ? 'fa-user' : 'fa-robot', 'text-2xl', msg.isUser ? 'text-blue-500' : 'text-green-500']"></i>
            </div>
            <div class="flex-grow">
              <p class="text-gray-800 text-left" v-html="markdownToHtml(msg.text)"></p>
              <i class="fas fa-spinner fa-1x fa-spin text-gray-800" v-if="!msg.text"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="flex-1 flex flex-col relative" style="height: 100%;">
        <!-- 图片上传按钮和预览区域 -->
        <div class="mb-2 px-4">
          <div class="flex items-center gap-2">
            <label for="image-upload" class="cursor-pointer inline-flex items-center text-gray-500 hover:text-gray-700">
              <i class="fas fa-image text-xl"></i>
              <span class="ml-2 text-sm">Add Image</span>
            </label>
            <input id="image-upload" type="file" accept="image/*" multiple class="hidden" @change="handleImageUpload">
            
            <!-- 显示上传状态和错误信息 -->
            <div v-if="isUploading" class="text-blue-500">
              <i class="fas fa-spinner fa-spin"></i>
              <span class="ml-2">Uploading...</span>
            </div>
            <div v-if="imageError" class="text-red-500 text-sm">
              {{ imageError }}
            </div>
          </div>

          <!-- 图片预览区域 -->
          <div class="mt-2 flex flex-wrap gap-2">
            <div v-for="image in uploadedImages" :key="image.id" 
                 class="relative group border border-gray-200 rounded-lg p-1">
              <img :src="image.src" 
                   :alt="image.name"
                   class="w-20 h-20 object-cover rounded-md cursor-pointer" 
                   @click="showFullImage(image.src)">
              <div class="absolute inset-0 bg-black bg-opacity-40 opacity-0 group-hover:opacity-100 
                          transition-opacity duration-200 rounded-md flex items-center justify-center">
                <button @click.stop="removeImage(image.id)" 
                        class="text-white hover:text-red-500 transition-colors duration-200">
                  <i class="fas fa-trash-alt"></i>
                </button>
              </div>
              <span class="absolute bottom-0 left-0 right-0 text-xs text-center bg-black bg-opacity-50 
                         text-white py-1 rounded-b-md">
                {{ (image.size / 1024).toFixed(1) }}KB
              </span>
            </div>
          </div>
        </div>

        <!-- 消息输入区域 -->
        <div class="flex-1 relative">
          <textarea v-model="userInput" placeholder="Type your question here..."
            class="w-full h-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none resize-none"></textarea>

          <!-- 浮动图标 -->
          <div class="floating-icons">
            <div class="absolute inset-y-0 right-0 flex items-center space-x-4 pr-4">
              <i class="fas fa-plus-square text-gray-500 cursor-pointer"></i>
              <i class="fas fa-file-alt text-gray-500 cursor-pointer"></i>
              <i class="fas fa-folder-open text-gray-500 cursor-pointer"></i>
              <i class="fas fa-paperclip text-gray-500 cursor-pointer"></i>
              <i :class="['fas', 'fa-paper-plane', 'text-blue-500', 'cursor-pointer', { 'opacity-50 cursor-not-allowed': !userInput.trim() }]"
                @click="sendMessage"></i>
            </div>
          </div>
        </div>

        <!-- 全尺寸图片模态框 -->
        <div v-if="fullImageSrc" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
          @click="fullImageSrc = null">
          <img :src="fullImageSrc" class="max-w-full max-h-full" @click.stop>
        </div>
      </div>
    </div>


    <!-- Settings Modal -->
    <div v-if="showSettings" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div class="bg-white rounded-lg p-6 w-96">
        <h2 class="text-2xl font-bold mb-4">Settings</h2>
        <form @submit.prevent="submitSettings">
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700">Name</label>
            <input v-model="settings.name" type="text"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50">
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700">Key</label>
            <input v-model="settings.key" type="text"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50">
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700">Base URL</label>
            <input v-model="settings.baseUrl" type="text"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50">
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700">Model</label>
            <input v-model="settings.model" type="text"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50">
          </div>
          <button type="submit" class="w-full bg-blue-500 text-white rounded-md py-2 hover:bg-blue-600">Submit</button>
        </form>
        <div class="mt-4">
          <h3 class="font-bold mb-2">Submitted Settings:</h3>
          <ul class="list-disc pl-5">
            <li v-for="(setting, index) in submittedSettings" :key="index" class="flex justify-between items-center">
              <span>Name: {{ setting.name }}, Key: {{ setting.key }}, Base URL: {{ setting.baseUrl }},Model:{{
                setting.model }}</span>
              <button @click="deleteSetting(index)" class="text-red-500 hover:text-red-700">
                <i class="fas fa-trash-alt"></i>
              </button>
            </li>
          </ul>
        </div>
        <button @click="showSettings = false"
          class="mt-4 w-full bg-gray-300 text-gray-800 rounded-md py-2 hover:bg-gray-400">Close</button>
      </div>
    </div>
    <!--Chat Setting -->
    <div v-if="showChatSetting" class="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50">
      <div class="bg-white p-6 rounded-lg w-96">
        <h2 class="text-lg font-bold mb-4">Settings</h2>
        <div class="mb-4">
          <label class="block text-sm font-bold mb-2">Title</label>
          <input v-model="selectTitle" type="text" class="w-full p-2 border border-gray-300 rounded">
        </div>
        <div class="mb-4">
          <label class="block text-sm font-bold mb-2">Model</label>
          <select v-model="selectedModel" class="w-full p-2 border border-gray-300 rounded">
            <option v-for="model in submittedSettings" :key="model.id" :value="model">{{ model.name }}</option>
          </select>
        </div>
        <div class="mb-4">
          <label class="block text-sm font-bold mb-2">Rounds</label>
          <input v-model="conversationRounds" type="number" class="w-full p-2 border border-gray-300 rounded">
        </div>
        <div class="mb-4">
          <label class="block text-sm font-bold mb-2">Input Tokens</label>
          <input v-model="maxInputTokens" type="number" class="w-full p-2 border border-gray-300 rounded">
        </div>
        <div class="mb-4">
          <label class="block text-sm font-bold mb-2">Output Tokens</label>
          <input v-model="maxOutputTokens" type="number" class="w-full p-2 border border-gray-300 rounded">
        </div>
        <div class="mb-4">
          <label class="block text-sm font-bold mb-2">System Prompt</label>
          <textarea v-model="systemPrompt" class="w-full p-2 border border-gray-300 rounded"></textarea>
        </div>
        <button @click="submitChatSettings" class="px-4 py-2 bg-blue-500 text-white rounded">Submit</button>
        <button @click="() => { showChatSetting = false }"
          class="px-4 py-2 bg-blue-500 text-white rounded">Close</button>
      </div>
    </div>

    <!-- About Model-->
    <div v-if="showAbout"
      class="fixed inset-0 mx-auto p-8 w-1/2 h-fit bg-gradient-to-r from-blue-100 to-purple-100 rounded-lg shadow-lg">
      <h1 class="text-3xl font-bold text-center text-blue-600 mb-6">欢迎来到ChatAlice的奇幻世界！🐰🍄</h1>

      <p class="text-lg text-gray-600 mb-3">
        ChatAlice 是一位充满好奇心的AI聊天伙伴，如同闯入科技奇境的现代爱丽丝。每次对话都是一场穿越想象力边界的冒险！
      </p>

      <p class="text-lg text-gray-600 mb-3">
        深奥的技术话题，天马行空的想象，Alice都会以温暖而智慧的方式陪伴左右。跟随我们的白兔，跳入这个充满惊喜的数字兔子洞吧？
      </p>

      <p class="text-lg text-gray-600 mb-6">
        有趣的是，在计算机世界里，Alice常常和的好朋友Bob一起出现在各种思想实验中。而在我们的项目里，Alice决定独自前行，成为你的专属对话伙伴！
      </p>

      <div class="text-center">
        <a href="https://github.com/JessonChan/chatalice"
          class="inline-block bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded transition duration-300">
          在GitHub上探索我们的仙境
        </a>
      </div>

      <p class="text-sm text-gray-500 mt-6 text-center italic">
        "在这里，我们都有点疯狂。你疯狂，我疯狂。但我会告诉你一个秘密，最棒的人都是疯狂的。" - 柴郡猫（可能是程序员）
      </p>
      <button @click="showAbout = false"
        class="mt-4 w-full bg-gray-300 text-gray-800 rounded-md py-2 hover:bg-gray-400">Close</button>
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
.message-scroll pre code{
  padding-top: 12px;
  padding-bottom: 32px;
  width: calc(70%);
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
.chat-title {
  display: -webkit-box; /* 使其成为弹性盒子 */
  -webkit-box-orient: vertical; /* 垂直排列 */
  -line-clamp: 2; /* 限制为两行 */
  overflow: hidden; /* 隐藏超出部分 */
  text-overflow: ellipsis; /* 使用省略号表示被隐藏的文本 */
}

.image-preview-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
  gap: 8px;
  padding: 8px;
}

.image-preview-item {
  position: relative;
  aspect-ratio: 1;
  overflow: hidden;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.image-preview-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-preview-item:hover .image-actions {
  opacity: 1;
}

.image-actions {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  opacity: 0;
  transition: opacity 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.upload-progress {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: #3b82f6;
  transition: width 0.3s ease;
}
</style>
